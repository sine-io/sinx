package agent

import (
	"crypto/tls"

	"github.com/hashicorp/serf/serf"
	"github.com/rs/zerolog"

	sxcfg "github.com/sine-io/sinx/internal/config"
)

// Logger returns the pointer to the agent's logger.
func (a *Agent) Logger() zerolog.Logger {
	return a.logger
}

// WithConfig option to set config to the agent
func (a *Agent) WithConfig(config *sxcfg.Config) *Agent {
	a.config = config

	return a
}

func (a *Agent) Config() *sxcfg.Config {
	return a.config
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

// WithStorage set the storage in the agent
func (a *Agent) WithStorage(s Storage) *Agent {
	a.Storage = s

	return a
}

// WithRaftStore set raft store in the agent
func (a *Agent) WithRaftStore(raftStore RaftStore) *Agent {
	a.raftStore = raftStore

	return a
}

func (a *Agent) WithSerf(s *serf.Serf) *Agent {
	a.serf = s

	return a
}

func (a *Agent) Serf() *serf.Serf {
	return a.serf
}

func (a *Agent) WithPeers(prs map[string][]*ServerParts) *Agent {
	a.peers = prs

	return a
}

func (a *Agent) Peers() map[string][]*ServerParts {
	return a.peers
}
