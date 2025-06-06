package cmd

import (
	"fmt"

	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"

	zlog "github.com/rs/zerolog/log"

	sxagent "github.com/sine-io/sinx/internal/agent"
	sxconfig "github.com/sine-io/sinx/internal/config"
)

var peerID string

func init() {
	raftCmd.PersistentFlags().StringVar(&rpcAddr, "rpc-addr", "{{ GetPrivateIP }}:6868", "gRPC address of the agent.")
	raftRemovePeerCmd.Flags().StringVar(&peerID, "peer-id", "", "Remove a Dkron server with the given ID from the Raft configuration.")

	raftCmd.AddCommand(raftListCmd)
	raftCmd.AddCommand(raftRemovePeerCmd)

	rootCmd.AddCommand(raftCmd)
}

// raftCmd represents the raft command
var raftCmd = &cobra.Command{
	Use:   "raft [command]",
	Short: "Command to perform some raft operations",
	Long:  ``,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ipa, err := sxconfig.ParseSingleIPTemplate(rpcAddr)
		if err != nil {
			return err
		}
		ip = ipa

		return nil
	},
}

var raftListCmd = &cobra.Command{
	Use:   "list-peers",
	Short: "Command to list raft peers",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		gc := sxagent.NewGRPCClient(nil, nil, zlog.Logger)

		reply, err := gc.RaftGetConfiguration(ip)
		if err != nil {
			return err
		}

		// Format it as a nice table.
		result := []string{"Node|ID|Address|State|Voter"}
		for _, s := range reply.Servers {
			state := "follower"
			if s.Leader {
				state = "leader"
			}
			result = append(result, fmt.Sprintf("%s|%s|%s|%s|%v",
				s.Node, s.Id, s.Address, state, s.Voter))
		}

		fmt.Println(columnize.SimpleFormat(result)) // TODO: use logger?

		return nil
	},
}

var raftRemovePeerCmd = &cobra.Command{
	Use:   "remove-peer",
	Short: "Command to remove a peer from raft",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		gc := sxagent.NewGRPCClient(nil, nil, zlog.Logger)

		if err := gc.RaftRemovePeerByID(ip, peerID); err != nil {
			return err
		}
		zlog.Info().Msg("Peer removed")

		return nil
	},
}
