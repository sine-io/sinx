package plugin

import (
	goplugin "github.com/hashicorp/go-plugin"
)

// The constants below are the names of the plugins that can be dispensed
// from the plugin server.
const (
	ProcessorPluginName = "processor"
	ExecutorPluginName  = "executor"
)

// Handshake is the HandshakeConfig used to configure clients and servers.
var Handshake = goplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "SINX_PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "b29a76488b6f3ca7955c5f769b50641f0fcd88748d8cedecda313d516320ca19",
}

// ServeOpts are the configurations to serve a plugin.
type ServeOpts struct {
	Processor Processor
	Executor  Executor
}

// Serve serves a plugin. This function never returns and should be the final
// function called in the main function of the plugin.
func Serve(opts *ServeOpts) {
	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins:         pluginMap(opts),
	})
}

// pluginMap returns the map[string]plugin.Plugin to use for configuring a plugin
// server or client.
func pluginMap(opts *ServeOpts) map[string]goplugin.Plugin {
	return map[string]goplugin.Plugin{
		"processor": &ProcessorPlugin{Processor: opts.Processor},
	}
}
