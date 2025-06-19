package agent

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/devopsfaith/krakend-usage/client"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	sxconfig "github.com/sine-io/sinx/internal/config"
	sxlog "github.com/sine-io/sinx/log"
	"github.com/soheilhy/cmux"
)

// StartServer launch a new SinX server process
func (a *Agent) StartServer() {
	if a.JobDB == nil {
		s, err := NewBuntJobDB()
		if err != nil {
			a.logger.Fatal().Err(err).Msg("sinx: Error initializing store")
		}

		// set store logger to zerolog
		a.JobDB = s.WithLogger(&a.logger)
	}

	// set schduler logger to zerolog
	a.sched = NewCronScheduler().WithLogger(&a.logger)

	// if a.HTTPTransport == nil {
	// 	a.HTTPTransport = NewHTTPTransport(a)
	// }
	a.HTTPTransport.ServeHTTP()

	// Create a cmux object.
	tcpm := cmux.New(a.listener)
	var grpcl, raftl net.Listener

	// If TLS config present listen to TLS
	if a.TLSConfig != nil {
		// Create a RaftLayer with TLS
		a.raftLayer = NewTLSRaftLayer(
			a.TLSConfig,
		)
		// set logger to raft layer
		a.raftLayer.WithLogger(&a.logger)

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
				a.logger.Fatal().Err(err)
			}
		}()
	} else {
		// Declare a plain RaftLayer
		a.raftLayer = NewRaftLayer()
		// set logger to raft layer
		a.raftLayer.WithLogger(&a.logger)

		// Declare the match for gRPC
		grpcl = tcpm.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))

		// Declare the match for raft RPC
		raftl = tcpm.Match(cmux.Any())
	}

	if a.GRPCServer == nil {
		// set gRPC server logger to zerolog
		a.GRPCServer = NewGRPCServer(a).WithLogger(&a.logger)
	}

	if err := a.GRPCServer.Serve(grpcl); err != nil {
		a.logger.Fatal().Err(err).Msg("agent: RPC server failed to start")
	}

	if err := a.raftLayer.Open(raftl); err != nil {
		a.logger.Fatal().Err(err).Send()
	}

	if err := a.setupRaft(); err != nil {
		a.logger.Fatal().Err(err).Msg("agent: Raft layer failed to start")
	}

	// Start serving everything
	go func() {
		if err := tcpm.Serve(); err != nil {
			a.logger.Fatal().Err(err).Send()
		}
	}()

	go a.monitorLeadership()

	a.startReporter()
}

func (a *Agent) setupRaft() error {
	if a.config.BootstrapExpect > 0 {
		if a.config.BootstrapExpect == 1 {
			a.config.Bootstrap = true
		}
	}

	transConfig := &raft.NetworkTransportConfig{
		Stream:                a.raftLayer,
		MaxPool:               3,
		Timeout:               raftTimeout,
		ServerAddressProvider: a.serverLookup,
		// set raft network logger to zerolog
		Logger: sxlog.HclogWrapper("raft-net", a.logger.GetLevel().String(), &a.logger),
	}
	transport := raft.NewNetworkTransportWithConfig(transConfig)
	a.raftTransport = transport

	raftCfg := raft.DefaultConfig()
	// set raft logger to zerolog
	raftCfg.Logger = sxlog.HclogWrapper("raft", a.logger.GetLevel().String(), &a.logger)

	// Raft performance
	raftMultiplier := a.config.RaftMultiplier
	if raftMultiplier < 1 || raftMultiplier > 10 {
		return fmt.Errorf("raft-multiplier cannot be %d. Must be between 1 and 10", raftMultiplier)
	}
	raftCfg.HeartbeatTimeout = raftCfg.HeartbeatTimeout * time.Duration(raftMultiplier)
	raftCfg.ElectionTimeout = raftCfg.ElectionTimeout * time.Duration(raftMultiplier)
	raftCfg.LeaderLeaseTimeout = raftCfg.LeaderLeaseTimeout * time.Duration(a.config.RaftMultiplier)

	raftCfg.LocalID = raft.ServerID(a.config.NodeName)

	// Build an all in-memory setup for dev mode, otherwise prepare a full
	// disk-based setup.
	var logStore raft.LogStore
	var stableStore raft.StableStore
	var snapshots raft.SnapshotStore
	if a.config.DevMode {
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
		snapshots, err = raft.NewFileSnapshotStoreWithLogger(
			filepath.Join(a.config.DataDir, "raft"), 3,
			sxlog.HclogWrapper("snapshot", a.logger.GetLevel().String(), &a.logger),
		)
		if err != nil {
			return fmt.Errorf("file snapshot store: %s", err)
		}

		// Create the BoltDB backend
		if a.raftStore == nil {
			s, err := raftboltdb.NewBoltStore(filepath.Join(a.config.DataDir, "raft", "raft.db"))
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
		peersFile := filepath.Join(a.config.DataDir, "raft", "peers.json")
		if _, err := os.Stat(peersFile); err == nil {
			a.logger.Info().Msg("found peers.json file, recovering Raft configuration...")

			var configuration raft.Configuration
			configuration, err = raft.ReadConfigJSON(peersFile)
			if err != nil {
				return fmt.Errorf("recovery failed to parse peers.json: %v", err)
			}

			store, err := NewBuntJobDB()
			if err != nil {
				a.logger.Fatal().Err(err).Msg("sinx: Error initializing store")
			}
			// set store logger to zerolog
			store.WithLogger(&a.logger)

			// set fsm logger to zerolog
			tmpFsm := newRaftFSM(store, nil).WithLogger(&a.logger)

			if err := raft.RecoverCluster(raftCfg, tmpFsm,
				logStore, stableStore, snapshots, transport, configuration); err != nil {
				return fmt.Errorf("recovery failed: %v", err)
			}
			if err := os.Remove(peersFile); err != nil {
				return fmt.Errorf("recovery failed to delete peers.json, please delete manually (see peers.info for details): %v", err)
			}
			a.logger.Info().Msg("deleted peers.json file after successful recovery")
		}
	}

	// If we are in bootstrap or dev mode and the state is clean then we can
	// bootstrap now.
	if a.config.Bootstrap || a.config.DevMode {
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
	fsm := newRaftFSM(a.JobDB, a.ProAppliers).WithLogger(&a.logger)

	rft, err := raft.NewRaft(raftCfg, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	a.leaderCh = rft.LeaderCh()
	a.raft = rft

	return nil
}

func (a *Agent) startReporter() {
	if a.config.DisableUsageStats || a.config.DevMode {
		a.logger.Info().Msg("agent: usage report client disabled")
		return
	}

	clusterID, err := a.config.Hash()
	if err != nil {
		a.logger.Warn().Msgf("agent: unable to hash the service configuration: %s", err.Error())
		return
	}

	go func() {
		serverID, _ := uuid.GenerateUUID()
		a.logger.Info().Msgf("agent: registering usage stats for cluster ID '%s'", clusterID)

		if err := client.StartReporter(context.Background(), client.Options{
			ClusterID: clusterID,
			ServerID:  serverID,
			URL:       "https://stats.xxxxxxx.io",
			Version:   fmt.Sprintf("%s %s", sxconfig.Name, sxconfig.Version),
		}); err != nil {
			a.logger.Warn().Msgf("agent: unable to create the usage report client: %s", err.Error())
		}
	}()
}
