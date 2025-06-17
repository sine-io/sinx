package flagset

import (
	flag "github.com/spf13/pflag"

	sxconfig "github.com/sine-io/sinx/internal/config"
)

// NodeFlagSet creates all of our node flags.
func NodeFlagSet(cfg *sxconfig.Config) *flag.FlagSet {
	cmdFlags := flag.NewFlagSet("node flagset", flag.ContinueOnError)

	cmdFlags.String("node-name", cfg.NodeName,
		"Name of this node. Must be unique in the cluster")

	cmdFlags.StringSlice("tag", []string{},
		`Tag can be specified multiple times to attach multiple key/value tag pairs 
to the given node, specified as key=value`)

	cmdFlags.String("datacenter", cfg.Datacenter,
		`Specifies the data center of the local agent. All members of a datacenter 
should share a local LAN connection.`)

	cmdFlags.String("region", cfg.Region,
		`Specifies the region the Dkron agent is a member of. A region typically maps 
to a geographic region, for example us, with potentially multiple zones, which 
map to datacenters such as us-west and us-east`)

	cmdFlags.Bool("server", false,
		"This node is running in server mode")

	cmdFlags.Bool("ui", true,
		"Enable the web UI on this node. The node must be server.")

	cmdFlags.String("profile", cfg.Profile,
		"Profile is used to control the timing profiles used")

	cmdFlags.String("data-dir", cfg.DataDir,
		`Specifies the directory to use for server-specific data, including the replicated log. 
		By default, this is the top-level data-dir, like [/var/lib/sinx]`)

	return cmdFlags
}
