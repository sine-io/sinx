package cmd

import (
	"github.com/hashicorp/go-plugin"
	"github.com/spf13/cobra"

	sxplugin "github.com/sine-io/sinx/plugin"
	sxhttp "github.com/sine-io/sinx/plugin/http"
)

func init() {
	rootCmd.AddCommand(httpCmd)
}

var httpCmd = &cobra.Command{
	Hidden: true,
	Use:    "http",
	Short:  "Run the http plugin",
	Long:   ``,
	Run: func(cmd *cobra.Command, args []string) {
		plugin.Serve(&plugin.ServeConfig{
			HandshakeConfig: sxplugin.Handshake,
			Plugins: map[string]plugin.Plugin{
				"executor": &sxplugin.ExecutorPlugin{Executor: sxhttp.New()},
			},

			// A non-nil value here enables gRPC serving for this plugin...
			GRPCServer: plugin.DefaultGRPCServer,
		})
	},
}
