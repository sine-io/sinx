package agent

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	sxconfig "github.com/sine-io/sinx/internal/config"
	sxproto "github.com/sine-io/sinx/types"
)

// StartAgent the current agent by running all the necessary
// checks and server or client routines.
func StartAgent(a *Agent) error {
	// Normalize configured addresses
	if err := a.Config.NormalizeAddrs(); err != nil && !errors.Is(err, sxconfig.ErrResolvingHost) {
		return err
	}

	s, err := a.setupSerf()
	if err != nil {
		return fmt.Errorf("agent: Can not setup serf, %s", err)
	}
	a.Serf = s

	// start retry join
	if len(a.Config.RetryJoinLAN) > 0 {
		a.retryJoinLAN()
	} else {
		_, err := a.join(a.Config.StartJoin, true)
		if err != nil {
			a.Logger.Warn().Err(err).Any("servers", a.Config.StartJoin).Msg("agent: Can not join")
		}
	}

	if err := initMetrics(a); err != nil {
		a.Logger.Fatal().Msg("agent: Can not setup metrics")
	}

	// Expose the node name
	expNode.Set(a.Config.NodeName)

	//Use the value of "RPCPort" if AdvertiseRPCPort has not been set
	if a.Config.AdvertiseRPCPort <= 0 {
		a.Config.AdvertiseRPCPort = a.Config.RPCPort
	}

	// Create a listener for RPC subsystem
	addr := a.bindRPCAddr()
	l, err := net.Listen("tcp", addr)
	if err != nil {
		a.Logger.Fatal().Err(err)
	}
	a.listener = l

	if a.Config.Server {
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
			if err := grpcServer.Serve(l); err != nil {
				a.Logger.Fatal().Err(err)
			}
		}()
	}

	if a.GRPCClient == nil {
		a.GRPCClient = NewGRPCClient(nil, a)
	}

	tags := a.Serf.LocalMember().Tags
	tags["rpc_addr"] = a.advertiseRPCAddr() // Address that clients will use to RPC to servers
	tags["port"] = strconv.Itoa(a.Config.AdvertiseRPCPort)
	if err := a.Serf.SetTags(tags); err != nil {
		return fmt.Errorf("agent: Error setting tags: %w", err)
	}

	go a.eventLoop()
	a.ready = true

	return nil
}
