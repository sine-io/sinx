package cmd

import (
	"fmt"

	"github.com/hashicorp/serf/serf"
	"github.com/spf13/cobra"

	sxagent "github.com/sine-io/sinx/internal/agent"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Long:  `Show the version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Name: %s\n", sxagent.Name)
		fmt.Printf("Version: %s\n", sxagent.Version)
		fmt.Printf("Codename: %s\n", sxagent.Codename)
		fmt.Printf("Agent Protocol: %d (Understands back to: %d)\n",
			serf.ProtocolVersionMax, serf.ProtocolVersionMin)
	},
}
