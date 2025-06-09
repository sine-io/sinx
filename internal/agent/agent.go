package agent

import (
	"context"
	"crypto/tls"
	"errors"
	"expvar"
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/devopsfaith/krakend-usage/client"
	metrics "github.com/hashicorp/go-metrics"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/hashicorp/serf/serf"
	"github.com/rs/zerolog"

	sxconfig "github.com/sine-io/sinx/internal/config"
	sxplugin "github.com/sine-io/sinx/plugin"
	sxproto "github.com/sine-io/sinx/types"
)

const (
	raftTimeout = 30 * time.Second
	// raftLogCacheSize is the maximum number of logs to cache in-memory.
	// This is used to reduce disk I/O for the recently committed entries.
	raftLogCacheSize = 512
	minRaftProtocol  = 3
)

var (
	expNode = expvar.NewString("node")

	// ErrLeaderNotFound is returned when obtained leader is not found in member list
	ErrLeaderNotFound = errors.New("no member leader found in member list")

	// ErrNoSuitableServer returns an error in case no suitable server to send the request is found.
	ErrNoSuitableServer = errors.New("no suitable server found to send the request, aborting")

	runningExecutions sync.Map
)

type RaftStore interface {
	raft.StableStore
	raft.LogStore
	Close() error
}

// Node is a shorter, more descriptive name for serf.Member
type Node = serf.Member

// Agent is the main struct that represents a SinX agent
type Agent struct {
	// ProcessorPlugins maps processor plugins
	ProcessorPlugins map[string]sxplugin.Processor

	//ExecutorPlugins maps executor plugins
	ExecutorPlugins map[string]sxplugin.Executor

	// HTTPTransport is a swappable interface for the HTTP server interface
	// HTTPTransport Transport

	// Store interface to set the storage engine
	Store Storage

	// GRPCServer interface for setting the GRPC server
	GRPCServer DkronGRPCServer

	// GRPCClient interface for setting the GRPC client
	GRPCClient DkronGRPCClient

	// TLSConfig allows setting a TLS config for transport
	TLSConfig *tls.Config

	// Pro features
	GlobalLock         bool
	MemberEventHandler func(serf.Event)
	ProAppliers        LogAppliers

	Serf        *serf.Serf
	eventCh     chan serf.Event
	sched       *Scheduler
	ready       bool
	shutdownCh  chan struct{}
	retryJoinCh chan error

	// The raft instance is used among SinX nodes within the
	// region to protect operations that require strong consistency
	leaderCh <-chan bool
	raft     *raft.Raft
	// raftLayer provides network layering of the raft RPC along with
	// the SinX gRPC transport layer.
	raftLayer     *RaftLayer
	raftStore     RaftStore
	raftInmem     *raft.InmemStore
	raftTransport *raft.NetworkTransport

	// reconcileCh is used to pass events from the serf handler
	// into the leader manager. Mostly used to handle when servers
	// join/leave from the region.
	reconcileCh chan serf.Member

	// peers is used to track the known SinX servers. This is
	// used for region forwarding and clustering.
	Peers        map[string][]*ServerParts
	localPeers   map[raft.ServerAddress]*ServerParts
	peerLock     sync.RWMutex
	serverLookup *ServerLookup

	activeExecutions sync.Map

	listener net.Listener

	// logger is the log entry to use for all logging calls
	Logger zerolog.Logger
	// config is the configuration to use for the agent
	Config *sxconfig.Config
}

// ProcessorFactory is a function type that creates a new instance
// of a processor.
type ProcessorFactory func() (sxplugin.Processor, error)

// Plugins struct to store loaded plugins of each type
type Plugins struct {
	Processors map[string]sxplugin.Processor
	Executors  map[string]sxplugin.Executor
}

// NewAgent returns a new Agent instance capable of starting
// and running a SinX instance.
func NewAgent(config *sxconfig.Config, options ...Options) *Agent {
	agent := &Agent{
		Config:       config,
		retryJoinCh:  make(chan error),
		serverLookup: NewServerLookup(),
	}

	for _, option := range options {
		option(agent)
	}

	return agent
}

// RetryJoinCh is a channel that transports errors
// from the retry join process.
func (a *Agent) RetryJoinCh() <-chan error {
	return a.retryJoinCh
}

// JoinLAN is used to have SinX join the inner-DC pool
// The target address should be another node inside the DC
// listening on the Serf LAN address
func (a *Agent) JoinLAN(addrs []string) (int, error) {
	return a.Serf.Join(addrs, true)
}

// UpdateTags updates the tag configuration for this agent
func (a *Agent) UpdateTags(tags map[string]string) {
	// Preserve reserved tags
	currentTags := a.Serf.LocalMember().Tags
	for _, tagName := range []string{"role", "version", "server", "bootstrap", "expect", "port", "rpc_addr"} {
		if val, exists := currentTags[tagName]; exists {
			tags[tagName] = val
		}
	}
	tags["dc"] = a.Config.Datacenter
	tags["region"] = a.Config.Region

	// Set new collection of tags
	err := a.Serf.SetTags(tags)
	if err != nil {
		a.Logger.Warn().Msgf("Setting tags unsuccessful: %s.", err.Error())
	}
}

func (a *Agent) setupRaft() error {
	if a.Config.BootstrapExpect > 0 {
		if a.Config.BootstrapExpect == 1 {
			a.Config.Bootstrap = true
		}
	}

	raftNetworkLogger := &a.Logger
	transConfig := &raft.NetworkTransportConfig{
		Stream:                a.raftLayer,
		MaxPool:               3,
		Timeout:               raftTimeout,
		ServerAddressProvider: a.serverLookup,
		// set raft network logger to zerolog
		Logger: customHclogWithZerolog("raft-net", a.Logger.GetLevel().String(), *raftNetworkLogger),
	}
	transport := raft.NewNetworkTransportWithConfig(transConfig)
	a.raftTransport = transport

	raftCfg := raft.DefaultConfig()
	// set raft logger to zerolog
	raftLogger := &a.Logger
	raftCfg.Logger = customHclogWithZerolog("raft", a.Logger.GetLevel().String(), *raftLogger)

	// Raft performance
	raftMultiplier := a.Config.RaftMultiplier
	if raftMultiplier < 1 || raftMultiplier > 10 {
		return fmt.Errorf("raft-multiplier cannot be %d. Must be between 1 and 10", raftMultiplier)
	}
	raftCfg.HeartbeatTimeout = raftCfg.HeartbeatTimeout * time.Duration(raftMultiplier)
	raftCfg.ElectionTimeout = raftCfg.ElectionTimeout * time.Duration(raftMultiplier)
	raftCfg.LeaderLeaseTimeout = raftCfg.LeaderLeaseTimeout * time.Duration(a.Config.RaftMultiplier)

	raftCfg.LocalID = raft.ServerID(a.Config.NodeName)

	// Build an all in-memory setup for dev mode, otherwise prepare a full
	// disk-based setup.
	var logStore raft.LogStore
	var stableStore raft.StableStore
	var snapshots raft.SnapshotStore
	if a.Config.DevMode {
		store := raft.NewInmemStore()
		a.raftInmem = store
		stableStore = store
		logStore = store
		snapshots = raft.NewDiscardSnapshotStore()
	} else {
		var err error
		// Create the snapshot store. This allows the Raft to truncate the log to
		// mitigate the issue of having an unbounded replicated log.
		// We set the snapshot logger to zerolog
		snapshotLogger := &a.Logger
		snapshots, err = raft.NewFileSnapshotStoreWithLogger(
			filepath.Join(a.Config.DataDir, "raft"), 3,
			customHclogWithZerolog("snapshot", a.Logger.GetLevel().String(), *snapshotLogger),
		)
		if err != nil {
			return fmt.Errorf("file snapshot store: %s", err)
		}

		// Create the BoltDB backend
		if a.raftStore == nil {
			s, err := raftboltdb.NewBoltStore(filepath.Join(a.Config.DataDir, "raft", "raft.db"))
			if err != nil {
				return fmt.Errorf("error creating new raft store: %s", err)
			}
			a.raftStore = s
		}
		stableStore = a.raftStore

		// Wrap the store in a LogCache to improve performance
		cacheStore, err := raft.NewLogCache(raftLogCacheSize, a.raftStore)
		if err != nil {
			a.raftStore.Close()
			return err
		}
		logStore = cacheStore

		// Check for peers.json file for recovery
		peersFile := filepath.Join(a.Config.DataDir, "raft", "peers.json")
		if _, err := os.Stat(peersFile); err == nil {
			a.Logger.Info().Msg("found peers.json file, recovering Raft configuration...")

			var configuration raft.Configuration
			configuration, err = raft.ReadConfigJSON(peersFile)
			if err != nil {
				return fmt.Errorf("recovery failed to parse peers.json: %v", err)
			}

			// set store logger to zerolog
			storeLogger := &a.Logger
			store, err := NewStore(storeLogger.Hook(
				zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
					e.Str("store-xxxxxx", msg)
				}),
			))
			if err != nil {
				a.Logger.Fatal().Err(err).Msg("sinx: Error initializing store")
			}

			// set fsm logger to zerolog
			tmpFsmLogger := &a.Logger
			tmpFsm := newRaftFSM(
				store, nil,
				tmpFsmLogger.Hook(
					zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
						e.Str("tmpFsm-xxxxxx", msg)
					}),
				))

			if err := raft.RecoverCluster(raftCfg, tmpFsm,
				logStore, stableStore, snapshots, transport, configuration); err != nil {
				return fmt.Errorf("recovery failed: %v", err)
			}
			if err := os.Remove(peersFile); err != nil {
				return fmt.Errorf("recovery failed to delete peers.json, please delete manually (see peers.info for details): %v", err)
			}
			a.Logger.Info().Msg("deleted peers.json file after successful recovery")
		}
	}

	// If we are in bootstrap or dev mode and the state is clean then we can
	// bootstrap now.
	if a.Config.Bootstrap || a.Config.DevMode {
		hasState, err := raft.HasExistingState(logStore, stableStore, snapshots)
		if err != nil {
			return err
		}
		if !hasState {
			configuration := raft.Configuration{
				Servers: []raft.Server{
					{
						ID:      raftCfg.LocalID,
						Address: transport.LocalAddr(),
					},
				},
			}
			if err := raft.BootstrapCluster(raftCfg, logStore, stableStore, snapshots, transport, configuration); err != nil {
				return err
			}
		}
	}

	// Instantiate the Raft systems. The second parameter is a finite state machine
	// which stores the actual kv pairs and is operated upon through Apply().
	// set fsm logger to zerolog.
	fsmLogger := &a.Logger
	fsm := newRaftFSM(
		a.Store, a.ProAppliers,
		fsmLogger.Hook(
			zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
				e.Str("fsm-xxxxxx", msg)
			}),
		),
	)

	rft, err := raft.NewRaft(raftCfg, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	a.leaderCh = rft.LeaderCh()
	a.raft = rft

	return nil
}

// setupSerf is used to create the agent we use
func (a *Agent) setupSerf() (*serf.Serf, error) {
	config := a.Config

	// Init peer list
	a.localPeers = make(map[raft.ServerAddress]*ServerParts)
	a.peers = make(map[string][]*ServerParts)

	bindIP, bindPort, err := config.AddrParts(config.BindAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid bind address: %s", err)
	}

	var advertiseIP string
	var advertisePort int
	if config.AdvertiseAddr != "" {
		advertiseIP, advertisePort, err = config.AddrParts(config.AdvertiseAddr)
		if err != nil {
			return nil, fmt.Errorf("invalid advertise address: %s", err)
		}
	}

	encryptKey, err := config.EncryptBytes()
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key: %s", err)
	}

	serfConfig := serf.DefaultConfig()
	serfConfig.Init()

	// set serf logger
	serfLogger := &a.Logger
	serfConfig.Logger = customGologWithZerolog(*serfLogger)

	serfConfig.Tags = a.Config.Tags
	serfConfig.Tags["role"] = "sinx"
	serfConfig.Tags["dc"] = a.Config.Datacenter
	serfConfig.Tags["region"] = a.Config.Region
	serfConfig.Tags["version"] = Version
	if a.Config.Server {
		serfConfig.Tags["server"] = strconv.FormatBool(a.Config.Server)
	}
	if a.Config.Bootstrap {
		serfConfig.Tags["bootstrap"] = "1"
	}
	if a.Config.BootstrapExpect != 0 {
		serfConfig.Tags["expect"] = fmt.Sprintf("%d", a.Config.BootstrapExpect)
	}

	switch config.Profile {
	case "lan":
		serfConfig.MemberlistConfig = memberlist.DefaultLANConfig()
	case "wan":
		serfConfig.MemberlistConfig = memberlist.DefaultWANConfig()
	case "local":
		serfConfig.MemberlistConfig = memberlist.DefaultLocalConfig()
	default:
		return nil, fmt.Errorf("unknown profile: %s", config.Profile)
	}

	serfConfig.MemberlistConfig.BindAddr = bindIP
	serfConfig.MemberlistConfig.BindPort = bindPort
	serfConfig.MemberlistConfig.AdvertiseAddr = advertiseIP
	serfConfig.MemberlistConfig.AdvertisePort = advertisePort
	serfConfig.MemberlistConfig.SecretKey = encryptKey

	// set serf memberlist logger
	serfMemberlistLogger := &a.Logger
	serfConfig.MemberlistConfig.Logger = customGologWithZerolog(*serfMemberlistLogger)

	serfConfig.NodeName = config.NodeName
	serfConfig.Tags = config.Tags
	serfConfig.CoalescePeriod = 3 * time.Second
	serfConfig.QuiescentPeriod = time.Second
	serfConfig.UserCoalescePeriod = 3 * time.Second
	serfConfig.UserQuiescentPeriod = time.Second

	serfConfig.ReconnectTimeout, err = time.ParseDuration(config.SerfReconnectTimeout)
	if err != nil {
		a.Logger.Fatal().Err(err).Send()
	}

	// Create a channel to listen for events from Serf
	a.eventCh = make(chan serf.Event, 2048)
	serfConfig.EventCh = a.eventCh

	// Start Serf
	a.Logger.Info().Msg("agent: SinX agent starting")

	// Create serf first
	serf, err := serf.Create(serfConfig)
	if err != nil {
		a.Logger.Error().Err(err).Send()
		return nil, err
	}
	return serf, nil
}

// Utility method to get leader nodename
func (a *Agent) LeaderMember() (*serf.Member, error) {
	l := a.raft.Leader()
	for _, member := range a.Serf.Members() {
		if member.Tags["rpc_addr"] == string(l) {
			return &member, nil
		}
	}
	return nil, ErrLeaderNotFound
}

// IsLeader checks if this server is the cluster leader
func (a *Agent) IsLeader() bool {
	return a.raft.State() == raft.Leader
}

// Members is used to return the members of the serf cluster
func (a *Agent) Members() []serf.Member {
	return a.Serf.Members()
}

// LocalMember is used to return the local node
func (a *Agent) LocalMember() serf.Member {
	return a.Serf.LocalMember()
}

// Leader is used to return the Raft leader
func (a *Agent) Leader() raft.ServerAddress {
	return a.raft.Leader()
}

// Servers returns a list of known server
func (a *Agent) Servers() (members []*ServerParts) {
	for _, member := range a.Serf.Members() {
		ok, parts := isServer(member)
		if !ok || member.Status != serf.StatusAlive {
			continue
		}
		members = append(members, parts)
	}
	return members
}

// LocalServers returns a list of the local known server
func (a *Agent) LocalServers() (members []*ServerParts) {
	for _, member := range a.Serf.Members() {
		ok, parts := isServer(member)
		if !ok || member.Status != serf.StatusAlive {
			continue
		}
		if a.Config.Region == parts.Region {
			members = append(members, parts)
		}
	}
	return members
}

// Listens to events from Serf and handle the event.
func (a *Agent) eventLoop() {
	serfShutdownCh := a.Serf.ShutdownCh()
	a.Logger.Info().Msg("agent: Listen for events")
	for {
		select {
		case e := <-a.eventCh:
			a.Logger.Info().Str("event", e.String()).Msg("agent: Received event")

			metrics.IncrCounter([]string{"agent", "event_received", e.String()}, 1)

			// Log all member events
			if me, ok := e.(serf.MemberEvent); ok {
				for _, member := range me.Members {
					a.Logger.Debug().
						Str("node", a.Config.NodeName).
						Str("member", member.Name).
						Any("event", e.EventType()).
						Msg("agent: Member event")
				}

				if a.MemberEventHandler != nil {
					a.MemberEventHandler(e)
				}

				// serfEventHandler is used to handle events from the serf cluster
				switch e.EventType() {
				case serf.EventMemberJoin:
					a.nodeJoin(me)
					a.localMemberEvent(me)
				case serf.EventMemberLeave, serf.EventMemberFailed:
					a.nodeFailed(me)
					a.localMemberEvent(me)
				case serf.EventMemberReap:
					a.localMemberEvent(me)
				case serf.EventMemberUpdate:
					a.lanNodeUpdate(me)
					a.localMemberEvent(me)
				case serf.EventUser, serf.EventQuery: // Ignore
				default:
					a.Logger.Warn().Str("event", e.String()).Msg("agent: Unhandled serf event")
				}
			}

		case <-serfShutdownCh:
			a.Logger.Warn().Msg("agent: Serf shutdown detected, quitting")

			return
		}
	}
}

// Join asks the Serf instance to join. See the Serf.Join function.
func (a *Agent) join(addrs []string, replay bool) (n int, err error) {
	a.Logger.Info().Msgf("agent: joining: %v replay: %v", addrs, replay)

	n, err = a.Serf.Join(addrs, !replay)
	if n > 0 {
		a.Logger.Info().Msgf("agent: joined: %d nodes", n)
	}
	if err != nil {
		a.Logger.Warn().Msgf("agent: error joining: %v", err)
	}

	return
}

func (a *Agent) getTargetNodes(tags map[string]string, selectFunc func([]Node) int) []Node {
	bareTags, cardinality := cleanTags(tags, a.Logger)
	nodes := a.getQualifyingNodes(a.Serf.Members(), bareTags)

	return selectNodes(nodes, cardinality, selectFunc)
}

// getQualifyingNodes returns all nodes in the cluster that are
// alive, in this agent's region and have all given tags
func (a *Agent) getQualifyingNodes(nodes []Node, bareTags map[string]string) []Node {
	// Determine the usable set of nodes
	qualifiers := filterArray(nodes, func(node Node) bool {
		return node.Status == serf.StatusAlive &&
			node.Tags["region"] == a.Config.Region &&
			nodeMatchesTags(node, bareTags)
	})

	return qualifiers
}

// The default selector function for getTargetNodes/selectNodes
func defaultSelector(nodes []Node) int {
	return rand.IntN(len(nodes)) // sine. use math/rand/v2
}

// selectNodes selects at most #cardinality from the given nodes using the selectFunc
func selectNodes(nodes []Node, cardinality int, selectFunc func([]Node) int) []Node {
	// Return all nodes immediately if they're all going to be selected
	numNodes := len(nodes)
	if numNodes <= cardinality {
		return nodes
	}

	for ; cardinality > 0; cardinality-- {
		// Select a node
		chosenIndex := selectFunc(nodes[:numNodes])

		// Swap picked node with the last one and reduce choices so it can't get picked again
		nodes[numNodes-1], nodes[chosenIndex] = nodes[chosenIndex], nodes[numNodes-1]
		numNodes--
	}

	return nodes[numNodes:]
}

// Returns all items from an array for which filterFunc returns true,
func filterArray(arr []Node, filterFunc func(Node) bool) []Node {
	for i := len(arr) - 1; i >= 0; i-- {
		if !filterFunc(arr[i]) {
			arr[i] = arr[len(arr)-1]
			arr = arr[:len(arr)-1]
		}
	}
	return arr
}

// This function is called when a client request the RPCAddress
// of the current member.
// in marathon, it would return the host's IP and advertise RPC port
func (a *Agent) advertiseRPCAddr() string {
	bindIP := a.Serf.LocalMember().Addr
	return net.JoinHostPort(bindIP.String(), strconv.Itoa(a.Config.AdvertiseRPCPort))
}

// Get bind address for RPC
func (a *Agent) bindRPCAddr() string {
	bindIP, _, _ := a.Config.AddrParts(a.Config.BindAddr)
	return net.JoinHostPort(bindIP, strconv.Itoa(a.Config.RPCPort))
}

// applySetJob is a helper method to be called when
// a job property need to be modified from the leader.
func (a *Agent) applySetJob(job *sxproto.Job) error {
	cmd, err := Encode(SetJobType, job)
	if err != nil {
		return err
	}
	af := a.raft.Apply(cmd, raftTimeout)
	if err := af.Error(); err != nil {
		return err
	}
	res := af.Response()
	switch res {
	case ErrParentJobNotFound:
		return ErrParentJobNotFound
	case ErrSameParent:
		return ErrParentJobNotFound
	}

	return nil
}

// RaftApply applies a command to the Raft log
func (a *Agent) RaftApply(cmd []byte) raft.ApplyFuture {
	return a.raft.Apply(cmd, raftTimeout)
}

// GetRunningJobs returns amount of active jobs of the local agent
func (a *Agent) GetRunningJobs() int {
	job := 0
	runningExecutions.Range(func(k, v interface{}) bool {
		job = job + 1
		return true
	})
	return job
}

// GetActiveExecutions returns running executions globally
func (a *Agent) GetActiveExecutions() ([]*sxproto.Execution, error) {
	var executions []*sxproto.Execution

	for _, s := range a.LocalServers() {
		exs, err := a.GRPCClient.GetActiveExecutions(s.RPCAddr.String())
		if err != nil {
			return nil, err
		}

		executions = append(executions, exs...)
	}

	return executions, nil
}

// CheckAndSelectServer Check if the server is alive and select it
func (a *Agent) CheckAndSelectServer() (string, error) {
	var peers []string
	for _, p := range a.LocalServers() {
		peers = append(peers, p.RPCAddr.String())
	}

	for _, peer := range peers {
		a.Logger.Debug().Str("peer", peer).Msg("Checking peer")
		conn, err := net.DialTimeout("tcp", peer, 1*time.Second)
		if err == nil {
			conn.Close()
			a.Logger.Debug().Str("peer", peer).Msg("Found good peer")

			return peer, nil
		}
	}
	return "", ErrNoSuitableServer
}

func (a *Agent) startReporter() {
	if a.Config.DisableUsageStats || a.Config.DevMode {
		a.Logger.Info().Msg("agent: usage report client disabled")
		return
	}

	clusterID, err := a.Config.Hash()
	if err != nil {
		a.Logger.Warn().Msgf("agent: unable to hash the service configuration: %s", err.Error())
		return
	}

	go func() {
		serverID, _ := uuid.GenerateUUID()
		a.Logger.Info().Msgf("agent: registering usage stats for cluster ID '%s'", clusterID)

		if err := client.StartReporter(context.Background(), client.Options{
			ClusterID: clusterID,
			ServerID:  serverID,
			URL:       "https://stats.xxxxxxx.io",
			Version:   fmt.Sprintf("%s %s", Name, Version),
		}); err != nil {
			a.Logger.Warn().Msgf("agent: unable to create the usage report client: %s", err.Error())
		}
	}()
}
