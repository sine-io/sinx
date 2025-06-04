package flagset

import (
	flag "github.com/spf13/pflag"

	"github.com/sine-io/sinx/internal/config"
)

// StorageFlagSet creates all of our storage flags.
func StorageFlagSet(cfg *config.Config) *flag.FlagSet {
	cmdFlags := flag.NewFlagSet("storage flagset", flag.ContinueOnError)

	cmdFlags.String("data-dir", cfg.DataDir,
		`Specifies the directory to use for server-specific data, including the 
replicated log. By default, this is the top-level data-dir, 
like [/var/lib/dkron]`)

	cmdFlags.Int("raft-multiplier", cfg.RaftMultiplier,
		`An integer multiplier used by servers to scale key Raft timing parameters.
Omitting this value or setting it to 0 uses default timing described below. 
Lower values are used to tighten timing and increase sensitivity while higher 
values relax timings and reduce sensitivity. Tuning this affects the time it 
takes to detect leader failures and to perform leader elections, at the expense 
of requiring more network and CPU resources for better performance. By default, 
Dkron will use a lower-performance timing that's suitable for minimal Dkron 
servers, currently equivalent to setting this to a value of 5 (this default 
may be changed in future versions of Dkron, depending if the target minimum 
server profile changes). Setting this to a value of 1 will configure Raft to 
its highest-performance mode is recommended for production Dkron servers. 
The maximum allowed value is 10.`)

	return cmdFlags
}
