package cmd

import (
	"fmt"

	"github.com/hashicorp/serf/serf"
	"github.com/spf13/cobra"
)

var (
	// Name store the name of this software
	Name = "SinX"
	// Version is the current version that will get replaced
	// on build.
	Version = "devel"
	// Codename codename of this series
	Codename = "Devel"
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
		fmt.Printf("Name: %s\n", Name)
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Codename: %s\n", Codename)
		fmt.Printf("Agent Protocol: %d (Understands back to: %d)\n",
			serf.ProtocolVersionMax, serf.ProtocolVersionMin)
	},
}
