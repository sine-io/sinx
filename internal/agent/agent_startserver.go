package agent

import (
	"crypto/tls"
	"net"

	"github.com/rs/zerolog"
	"github.com/soheilhy/cmux"
)

// StartServer launch a new SinX server process
func (a *Agent) StartServer() {
	if a.Store == nil {
		// set store logger to zerolog
		storeLogger := &a.Logger
		s, err := NewStore(
			storeLogger.Hook(
				zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
					e.Str("store-xxxxxx", msg)
				}),
			),
		)
		if err != nil {
			a.Logger.Fatal().Err(err).Msg("sinx: Error initializing store")
		}

		a.Store = s
	}

	// set schduler logger to zerolog
	schdLogger := &a.Logger
	a.sched = NewScheduler(
		schdLogger.Hook(),
	)

	if a.HTTPTransport == nil {
		a.HTTPTransport = NewHTTPTransport(a)
	}
	a.HTTPTransport.ServeHTTP()

	// Create a cmux object.
	tcpm := cmux.New(a.listener)
	var grpcl, raftl net.Listener

	// set raft layer logger to zerolog
	raftLayerLogger := &a.Logger
	// If TLS config present listen to TLS
	if a.TLSConfig != nil {
		// Create a RaftLayer with TLS
		a.raftLayer = NewTLSRaftLayer(
			a.TLSConfig,
			raftLayerLogger.Hook(
				zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
					e.Str("tls-raft-layer-xxxxxx", msg)
				}),
			),
		)

		// Match any connection to the recursive mux
		tlsl := tcpm.Match(cmux.Any())
		tlsl = tls.NewListener(tlsl, a.TLSConfig)

		// Declare sub cMUX for TLS
		tlsm := cmux.New(tlsl)

		// Declare the match for TLS gRPC
		grpcl = tlsm.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))

		// Declare the match for TLS raft RPC
		raftl = tlsm.Match(cmux.Any())

		go func() {
			if err := tlsm.Serve(); err != nil {
				a.Logger.Fatal().Err(err)
			}
		}()
	} else {
		// Declare a plain RaftLayer
		a.raftLayer = NewRaftLayer(
			raftLayerLogger.Hook(
				zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
					e.Str("raft-layer-xxxxxx", msg)
				}),
			),
		)

		// Declare the match for gRPC
		grpcl = tcpm.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))

		// Declare the match for raft RPC
		raftl = tcpm.Match(cmux.Any())
	}

	if a.GRPCServer == nil {
		// set gRPC server logger to zerolog
		grpcLogger := &a.Logger
		a.GRPCServer = NewGRPCServer(
			a,
			grpcLogger.Hook(
				zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
					e.Str("grpc-server-xxxxxx", msg)
				}),
			),
		)
	}

	if err := a.GRPCServer.Serve(grpcl); err != nil {
		a.Logger.Fatal().Err(err).Msg("agent: RPC server failed to start")
	}

	if err := a.raftLayer.Open(raftl); err != nil {
		a.Logger.Fatal().Err(err).Send()
	}

	if err := a.setupRaft(); err != nil {
		a.Logger.Fatal().Err(err).Msg("agent: Raft layer failed to start")
	}

	// Start serving everything
	go func() {
		if err := tcpm.Serve(); err != nil {
			a.Logger.Fatal().Err(err).Send()
		}
	}()
	go a.monitorLeadership()
	a.startReporter()
}
