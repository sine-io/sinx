package agent

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	sconfig "github.com/sine-io/sinx/internal/config"
	srpc "github.com/sine-io/sinx/internal/rpc"
	slogging "github.com/sine-io/sinx/logging"
	sproto "github.com/sine-io/sinx/types"
)

// Start the current agent by running all the necessary
// checks and server or client routines.
func (a *Agent) Start() error {
	a.logger = slogging.L // Use the global logger, because L is singleton.

	// Initialize rand with current time
	// rand.Seed(time.Now().UnixNano()) // sine.2025.5.29, use math/rand/v2

	// Normalize configured addresses
	if err := a.config.NormalizeAddrs(); err != nil && !errors.Is(err, sconfig.ErrResolvingHost) {
		return err
	}

	s, err := a.setupSerf()
	if err != nil {
		return fmt.Errorf("agent: Can not setup serf, %s", err)
	}
	a.serf = s

	// start retry join
	if len(a.config.RetryJoinLAN) > 0 {
		a.retryJoinLAN()
	} else {
		_, err := a.join(a.config.StartJoin, true)
		if err != nil {
			a.logger.Warn().Err(err).Any("servers", a.config.StartJoin).Msg("agent: Can not join")
		}
	}

	if err := initMetrics(a); err != nil {
		a.logger.Fatal().Msg("agent: Can not setup metrics")
	}

	// Expose the node name
	expNode.Set(a.config.NodeName)

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

	if a.config.Server {
		a.StartServer()
	} else {
		var opts []grpc.ServerOption
		if a.TLSConfig != nil {
			tc := credentials.NewTLS(a.TLSConfig)
			opts = append(opts, grpc.Creds(tc))
		}

		grpcServer := grpc.NewServer(opts...)
		as := srpc.NewAgentServer(a, a.logger)
		sproto.RegisterAgentServer(grpcServer, as)
		go func() {
			if err := grpcServer.Serve(l); err != nil {
				a.logger.Fatal().Err(err)
			}
		}()
	}

	if a.GRPCClient == nil {
		a.GRPCClient = srpc.NewGRPCClient(nil, a, a.logger)
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
