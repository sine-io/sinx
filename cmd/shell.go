package cmd

import (
	"github.com/hashicorp/go-plugin"
	dkplugin "github.com/sine-io/sinx/plugin"
	"github.com/sine-io/sinx/plugin/shell"
	"github.com/spf13/cobra"
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
			HandshakeConfig: dkplugin.Handshake,
			Plugins: map[string]plugin.Plugin{
				"executor": &dkplugin.ExecutorPlugin{Executor: &shell.Shell{}},
			},

			// A non-nil value here enables gRPC serving for this plugin...
			GRPCServer: plugin.DefaultGRPCServer,
		})
	},
}
