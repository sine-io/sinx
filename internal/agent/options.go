package agent

import (
	"crypto/tls"
)

// Options type that defines agent options
type Options func(agent *Agent)

// WithPlugins option to set plugins to the agent
func WithPlugins(plugins Plugins) Options {
	return func(agent *Agent) {
		agent.ProcessorPlugins = plugins.Processors
		agent.ExecutorPlugins = plugins.Executors
	}
}

// WithTransportCredentials set tls config in the agent
func WithTransportCredentials(tls *tls.Config) Options {
	return func(agent *Agent) {
		agent.TLSConfig = tls
	}
}

// WithStore set store in the agent
func WithStore(store Storage) Options {
	return func(agent *Agent) {
		agent.Store = store
	}
}

func WithRaftStore(raftStore RaftStore) Options {
	return func(agent *Agent) {
		agent.raftStore = raftStore
	}
}
