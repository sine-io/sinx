package agent

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"

	sxconfig "github.com/sine-io/sinx/internal/config"
	sxmetrics "github.com/sine-io/sinx/internal/metrics"
	sxlog "github.com/sine-io/sinx/log"
	sxproto "github.com/sine-io/sinx/types"
)

// StartAgent the current agent by running all the necessary
// checks and server or client routines.
func (a *Agent) StartAgent() error {
	// 1. Normalize configured addresses
	if err := a.config.NormalizeAddrs(); err != nil && !errors.Is(err, sxconfig.ErrResolvingHost) {
		return err
	}

	// 2. Setup the cluster
	//    This will setup serf, retry-join and other cluster related things.
	if err := a.setupCluster(); err != nil {
		return err
	}

	// 3. Setup metrics
	//    This will setup the metrics sinks and start the metrics server.
	if err := sxmetrics.SetupMetrics(a.config); err != nil {
		a.logger.Fatal().Msg("agent: Can not setup metrics")
	}

	// 4. Expose the node name
	expNode.Set(a.config.NodeName)

	// 5. Setup network
	if err := a.setupNetwork(); err != nil {
		return err
	}

	if a.config.Server {
		a.StartServer()
	} else {
		var opts []grpc.ServerOption
		if a.TLSConfig != nil {
			tc := credentials.NewTLS(a.TLSConfig)
			opts = append(opts, grpc.Creds(tc))
		}

		grpcServer := grpc.NewServer(opts...)
		as := NewGRPCAgentServer(a)
		sxproto.RegisterAgentServer(grpcServer, as)
		go func() {
			if err := grpcServer.Serve(a.listener); err != nil {
				a.logger.Fatal().Err(err)
			}
		}()
	}

	if a.GRPCClient == nil {
		a.GRPCClient = NewGRPCClient(nil, a)
	}

	tags := a.serf.LocalMember().Tags
	tags["rpc_addr"] = a.advertiseRPCAddr() // Address that clients will use to RPC to servers
	tags["port"] = strconv.Itoa(a.config.AdvertiseRPCPort)
	if err := a.serf.SetTags(tags); err != nil {
		return fmt.Errorf("agent: Error setting tags: %w", err)
	}

	go a.eventLoop()
	a.ready = true

	return nil
}

func (a *Agent) setupCluster() error {
	// setup serf
	s, err := a.setupSerf()
	if err != nil {
		return fmt.Errorf("agent: Can not setup serf, %s", err)
	}
	a.serf = s

	// start retry join
	if len(a.config.RetryJoinLAN) > 0 {
		a.retryJoinLAN()
	} else {
		// join the existing serf cluster via serf.Join
		_, err := a.joinSerfCluster(a.config.StartJoin, true)
		if err != nil {
			a.logger.Warn().Err(err).Any("servers", a.config.StartJoin).Msg("agent: Can not join")
		}
	}

	return nil
}

// setupSerf is used to create the agent we use
func (a *Agent) setupSerf() (*serf.Serf, error) {
	config := a.config

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
	serfConfig.Logger = sxlog.GologWrapper(&a.logger)

	serfConfig.Tags = a.config.Tags
	serfConfig.Tags["role"] = "sinx"
	serfConfig.Tags["dc"] = a.config.Datacenter
	serfConfig.Tags["region"] = a.config.Region
	serfConfig.Tags["version"] = sxconfig.Version
	if a.config.Server {
		serfConfig.Tags["server"] = strconv.FormatBool(a.config.Server)
	}
	if a.config.Bootstrap {
		serfConfig.Tags["bootstrap"] = "1"
	}
	if a.config.BootstrapExpect != 0 {
		serfConfig.Tags["expect"] = fmt.Sprintf("%d", a.config.BootstrapExpect)
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
	serfConfig.MemberlistConfig.Logger = sxlog.GologWrapper(&a.logger)

	serfConfig.NodeName = config.NodeName
	serfConfig.Tags = config.Tags
	serfConfig.CoalescePeriod = 3 * time.Second
	serfConfig.QuiescentPeriod = time.Second
	serfConfig.UserCoalescePeriod = 3 * time.Second
	serfConfig.UserQuiescentPeriod = time.Second

	serfConfig.ReconnectTimeout, err = time.ParseDuration(config.SerfReconnectTimeout)
	if err != nil {
		a.logger.Fatal().Err(err).Send()
	}

	// Create a channel to listen for events from Serf
	a.eventCh = make(chan serf.Event, 2048)
	serfConfig.EventCh = a.eventCh

	// Start Serf
	a.logger.Info().Msg("agent: SinX agent starting")

	// Create serf first
	serf, err := serf.Create(serfConfig)
	if err != nil {
		a.logger.Error().Err(err).Send()
		return nil, err
	}
	return serf, nil
}

// joinSerfCluster asks the Serf instance to join. See the Serf.Join function.
func (a *Agent) joinSerfCluster(addrs []string, replay bool) (n int, err error) {
	a.logger.Info().Msgf("agent: joining: %v replay: %v", addrs, replay)

	n, err = a.serf.Join(addrs, !replay)
	if n > 0 {
		a.logger.Info().Msgf("agent: joined: %d nodes", n)
	}
	if err != nil {
		a.logger.Warn().Msgf("agent: error joining: %v", err)
	}

	return
}

func (a *Agent) setupNetwork() error {
	//Use the value of "RPCPort" if AdvertiseRPCPort has not been set
	if a.config.AdvertiseRPCPort <= 0 {
		a.config.AdvertiseRPCPort = a.config.RPCPort
	}

	// Create a listener for RPC subsystem
	addr := a.bindRPCAddr()
	l, err := net.Listen("tcp", addr)
	if err != nil {
		a.logger.Fatal().Err(err)
	}
	a.listener = l

	return nil
}

// Get bind address for RPC
func (a *Agent) bindRPCAddr() string {
	bindIP, _, _ := a.config.AddrParts(a.config.BindAddr)

	return net.JoinHostPort(bindIP, strconv.Itoa(a.config.RPCPort))
}
