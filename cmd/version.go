package cmd

import (
	"fmt"

	"github.com/hashicorp/serf/serf"
	"github.com/spf13/cobra"

	sxcfg "github.com/sine-io/sinx/internal/config"
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
		fmt.Printf("Name: %s\n", sxcfg.Name)
		fmt.Printf("Version: %s\n", sxcfg.Version)
		fmt.Printf("Codename: %s\n", sxcfg.Codename)
		fmt.Printf("Agent Protocol: %d (Understands back to: %d)\n",
			serf.ProtocolVersionMax, serf.ProtocolVersionMin)
	},
}
