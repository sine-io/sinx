package agent

import (
	"crypto/tls"

	"github.com/hashicorp/serf/serf"
	"github.com/rs/zerolog"

	sxconfig "github.com/sine-io/sinx/internal/config"
)

// WithLogger option to set logger to the agent
func (a *Agent) WithLogger(logger *zerolog.Logger) *Agent {
	a.logger = logger.Hook(
		zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
			e.Str("node", a.config.NodeName) // Add node name to each log event
		}),
	)

	return a
}

// Logger returns the logger struct
func (a *Agent) Logger() zerolog.Logger {
	return a.logger
}

// WithConfig option to set config to the agent
func (a *Agent) WithConfig(config *sxconfig.Config) *Agent {
	a.config = config

	return a
}

// WithPlugins option to set plugins to the agent
func (a *Agent) WithPlugins(plugins PluginRegistry) *Agent {
	a.ProcessorPlugins = plugins.Processors
	a.ExecutorPlugins = plugins.Executors

	return a
}

// WithTransportCredentials set tls config in the agent
func (a *Agent) WithTransportCredentials(tls *tls.Config) *Agent {
	a.TLSConfig = tls

	return a
}

// WithJobDB set job db in the agent
func (a *Agent) WithJobDB(jdb JobDB) *Agent {
	a.JobDB = jdb

	return a
}

// WithRaftStore set raft store in the agent
func (a *Agent) WithRaftStore(raftStore RaftStore) *Agent {
	a.raftStore = raftStore

	return a
}

func (a *Agent) WithSerf(s *serf.Serf) *Agent {
	a.Serf = s

	return a
}
