package main

import (
	"github.com/hashicorp/go-plugin"

	sxplugin "github.com/sine-io/sinx/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: sxplugin.Handshake,
		Plugins: map[string]plugin.Plugin{
			"executor": &sxplugin.ExecutorPlugin{Executor: &RabbitMQ{}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
