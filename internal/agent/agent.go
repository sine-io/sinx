package agent

import (
	"crypto/tls"
	"errors"
	"expvar"
	"net"
	"strconv"
	"sync"
	"time"

	metrics "github.com/hashicorp/go-metrics"
	"github.com/hashicorp/raft"
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

// Node is a shorter, more descriptive name for serf.Member
type Node = serf.Member

// Agent is the main struct that represents a SinX agent
type Agent struct {
	// ProcessorPlugins maps processor plugins
	ProcessorPlugins map[string]sxplugin.Processor

	//ExecutorPlugins maps executor plugins
	ExecutorPlugins map[string]sxplugin.Executor

	// HTTPTransport is a swappable interface for the HTTP server interface
	HTTPTransport Transport

	// JobDB interface to set the job db engine
	JobDB JobDB

	// GRPCServer interface for setting the GRPC server
	GRPCServer SinxGRPCServer

	// GRPCClient interface for setting the GRPC client
	GRPCClient SinxGRPCClient

	// TLSConfig allows setting a TLS config for transport
	TLSConfig *tls.Config

	// Pro features
	GlobalLock         bool
	MemberEventHandler func(serf.Event)
	ProAppliers        LogAppliers

	serf        *serf.Serf
	eventCh     chan serf.Event
	sched       Scheduler
	ready       bool
	shutdownCh  chan struct{}
	retryJoinCh chan error

	// The raft instance is used among SinX nodes within the
	// region to protect operations that require strong consistency
	leaderCh <-chan bool
	raft     *raft.Raft
	// raftLayer provides network layering of the raft RPC along with
	// the SinX gRPC transport layer.
	raftLayer *RaftLayer
	// raftStore is the store used to persist raft logs and state.
	raftStore RaftStore
	// raftInmem is the in-memory store used for development mode.
	raftInmem *raft.InmemStore
	// raftTransport is the network transport used by raft to communicate
	raftTransport *raft.NetworkTransport

	// reconcileCh is used to pass events from the serf handler
	// into the leader manager. Mostly used to handle when servers
	// join/leave from the region.
	reconcileCh chan serf.Member

	// peers is used to track the known SinX servers. This is
	// used for region forwarding and clustering.
	peers        map[string][]*ServerParts
	localPeers   map[raft.ServerAddress]*ServerParts
	peerLock     sync.RWMutex
	serverLookup *ServerLookup

	activeExecutions sync.Map

	listener net.Listener

	// logger is the log entry to use for all logging calls
	logger zerolog.Logger
	// config is the configuration to use for the agent
	config *sxconfig.Config
}

// ProcessorFactory is a function type that creates a new instance
// of a processor.
type ProcessorFactory func() (sxplugin.Processor, error)

// PluginRegistry struct to store loaded plugins of each type
type PluginRegistry struct {
	Processors map[string]sxplugin.Processor
	Executors  map[string]sxplugin.Executor
}

// NewAgent returns a new Agent instance capable of starting
// and running a SinX instance.
func NewAgent(config *sxconfig.Config) *Agent {
	agent := &Agent{
		config:       config,
		retryJoinCh:  make(chan error),
		serverLookup: NewServerLookup(),

		// set default logger, you can override it with WithLogger
		logger: zerolog.New(zerolog.NewConsoleWriter()),
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
	return a.serf.Join(addrs, true)
}

// UpdateTags updates the tag configuration for this agent
func (a *Agent) UpdateTags(tags map[string]string) {
	// Preserve reserved tags
	currentTags := a.serf.LocalMember().Tags
	for _, tagName := range []string{"role", "version", "server", "bootstrap", "expect", "port", "rpc_addr"} {
		if val, exists := currentTags[tagName]; exists {
			tags[tagName] = val
		}
	}
	tags["dc"] = a.config.Datacenter
	tags["region"] = a.config.Region

	// Set new collection of tags
	err := a.serf.SetTags(tags)
	if err != nil {
		a.logger.Warn().Msgf("Setting tags unsuccessful: %s.", err.Error())
	}
}

// Utility method to get leader nodename
func (a *Agent) LeaderMember() (*serf.Member, error) {
	l := a.raft.Leader()
	for _, member := range a.serf.Members() {
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
	return a.serf.Members()
}

// LocalMember is used to return the local node
func (a *Agent) LocalMember() serf.Member {
	return a.serf.LocalMember()
}

// Leader is used to return the Raft leader
func (a *Agent) Leader() raft.ServerAddress {
	return a.raft.Leader()
}

// Servers returns a list of known server
func (a *Agent) Servers() (members []*ServerParts) {
	for _, member := range a.serf.Members() {
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
	for _, member := range a.serf.Members() {
		ok, parts := isServer(member)
		if !ok || member.Status != serf.StatusAlive {
			continue
		}
		if a.config.Region == parts.Region {
			members = append(members, parts)
		}
	}
	return members
}

// Listens to events from Serf and handle the event.
func (a *Agent) eventLoop() {
	serfShutdownCh := a.serf.ShutdownCh()
	a.logger.Info().Msg("agent: Listen for events")
	for {
		select {
		case e := <-a.eventCh:
			a.logger.Info().Str("event", e.String()).Msg("agent: Received event")

			metrics.IncrCounter([]string{"agent", "event_received", e.String()}, 1)

			// Log all member events
			if me, ok := e.(serf.MemberEvent); ok {
				for _, member := range me.Members {
					a.logger.Debug().
						Str("node", a.config.NodeName).
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
					a.logger.Warn().Str("event", e.String()).Msg("agent: Unhandled serf event")
				}
			}

		case <-serfShutdownCh:
			a.logger.Warn().Msg("agent: Serf shutdown detected, quitting")

			return
		}
	}
}

// This function is called when a client request the RPCAddress
// of the current member.
// in marathon, it would return the host's IP and advertise RPC port
func (a *Agent) advertiseRPCAddr() string {
	bindIP := a.serf.LocalMember().Addr
	return net.JoinHostPort(bindIP.String(), strconv.Itoa(a.config.AdvertiseRPCPort))
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
	runningExecutions.Range(func(k, v any) bool {
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
		a.logger.Debug().Str("peer", peer).Msg("Checking peer")
		conn, err := net.DialTimeout("tcp", peer, 1*time.Second)
		if err == nil {
			conn.Close()
			a.logger.Debug().Str("peer", peer).Msg("Found good peer")

			return peer, nil
		}
	}
	return "", ErrNoSuitableServer
}
