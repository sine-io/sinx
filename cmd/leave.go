package cmd

import (
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	sxagent "github.com/sine-io/sinx/internal/agent"
	sxcfg "github.com/sine-io/sinx/internal/config"
)

func init() {
	rootCmd.AddCommand(leaveCmd)
	leaveCmd.PersistentFlags().StringVar(&rpcAddr, "rpc-addr", "{{ GetPrivateIP }}:6868", "gRPC address of the agent")
}

// versionCmd represents the version command
var leaveCmd = &cobra.Command{
	Use:   "leave",
	Short: "Force an agent to leave the cluster",
	Long: `Stop stops an agent, if the agent is a server and is running for election
	stop running for election, if this server was the leader
	this will force the cluster to elect a new leader and start a new scheduler.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		ipa, err := sxcfg.ParseSingleIPTemplate(rpcAddr)
		if err != nil {
			return err
		}
		ip = ipa

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		gc := sxagent.NewGRPCClient(nil, nil)

		if err := gc.Leave(ip); err != nil {
			return err
		}

		zlog.Info().Msg("Left the cluster successfully")
		return nil
	},
}
