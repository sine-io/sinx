package cmd

import (
	"github.com/spf13/cobra"

	sxconfig "github.com/sine-io/sinx/internal/config"
	sxrpc "github.com/sine-io/sinx/internal/rpc"
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
		ipa, err := sxconfig.ParseSingleIPTemplate(rpcAddr)
		if err != nil {
			return err
		}
		ip = ipa

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var gc sxrpc.DkronGRPCClient
		gc = sxrpc.NewGRPCClient(nil, nil, logger)

		if err := gc.Leave(ip); err != nil {
			return err
		}

		logger.Info().Msg("Left the cluster successfully")
		return nil
	},
}
