package cmd

import (
	"github.com/hashicorp/go-plugin"
	"github.com/spf13/cobra"

	sxplugin "github.com/sine-io/sinx/plugin"
	sxshell "github.com/sine-io/sinx/plugin/shell"
)

func init() {
	rootCmd.AddCommand(shellCmd)
}

var shellCmd = &cobra.Command{
	Hidden: true,
	Use:    "shell",
	Short:  "Shell plugin for dkron",
	Long:   ``,
	Run: func(cmd *cobra.Command, args []string) {
		plugin.Serve(&plugin.ServeConfig{
			HandshakeConfig: sxplugin.Handshake,
			Plugins: map[string]plugin.Plugin{
				"executor": &sxplugin.ExecutorPlugin{Executor: &sxshell.Shell{}},
			},

			// A non-nil value here enables gRPC serving for this plugin...
			GRPCServer: plugin.DefaultGRPCServer,
		})
	},
}
