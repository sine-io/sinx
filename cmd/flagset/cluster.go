package flagset

import (
	flag "github.com/spf13/pflag"

	sxconfig "github.com/sine-io/sinx/internal/config"
)

// ClusterFlagSet creates all of our cluster flags.
func ClusterFlagSet(cfg *sxconfig.Config) *flag.FlagSet {
	cmdFlags := flag.NewFlagSet("cluster flagset", flag.ContinueOnError)

	cmdFlags.String("encrypt", "",
		"Key for encrypting network traffic. Must be a base64-encoded 16-byte key")

	cmdFlags.StringSlice("join", []string{},
		"An initial agent to join with. This flag can be specified multiple times")

	cmdFlags.StringSlice("retry-join", []string{},
		`Address of an agent to join at start time with retries enabled. 
Can be specified multiple times.`)

	cmdFlags.Int("retry-max", 0,
		`Maximum number of join attempts. Defaults to 0, which will retry indefinitely.`)
	cmdFlags.String("retry-interval", cfg.RetryJoinIntervalLAN.String(),
		"Time to wait between join attempts.")

	cmdFlags.Int("bootstrap-expect", 0,
		`Provides the number of expected servers in the datacenter. Either this value 
should not be provided or the value must agree with other servers in the 
cluster. When provided, Dkron waits until the specified number of servers are 
available and then bootstraps the cluster. This allows an initial leader to be 
elected automatically. This flag requires server mode.`)

	return cmdFlags
}
