package cmd

import (
	"github.com/hashicorp/go-plugin"
	dkplugin "github.com/sine-io/sinx/plugin"
	"github.com/sine-io/sinx/plugin/http"
	"github.com/spf13/cobra"
)

var httpCmd = &cobra.Command{
	Hidden: true,
	Use:    "http",
	Short:  "Run the http plugin",
	Long:   ``,
	Run: func(cmd *cobra.Command, args []string) {
		plugin.Serve(&plugin.ServeConfig{
			HandshakeConfig: dkplugin.Handshake,
			Plugins: map[string]plugin.Plugin{
				"executor": &dkplugin.ExecutorPlugin{Executor: http.New()},
			},

			// A non-nil value here enables gRPC serving for this plugin...
			GRPCServer: plugin.DefaultGRPCServer,
		})
	},
}

func init() {
	dkronCmd.AddCommand(httpCmd)
}
