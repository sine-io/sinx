package main

import (
	dkplugin "github.com/distribworks/dkron/v4/plugin"
	"github.com/hashicorp/go-plugin"
)

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: dkplugin.Handshake,
		Plugins: map[string]plugin.Plugin{
			"executor": &dkplugin.ExecutorPlugin{Executor: &S3{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
