package flagset

import (
	flag "github.com/spf13/pflag"

	sxcfg "github.com/sine-io/sinx/internal/config"
)

// ObservabilityFlagSet creates all of our observability flags.
func ObservabilityFlagSet(cfg *sxcfg.Config) *flag.FlagSet {
	cmdFlags := flag.NewFlagSet("observability flagset", flag.ContinueOnError)

	cmdFlags.String("dog-statsd-addr", "", "DataDog Agent address")
	cmdFlags.StringSlice("dog-statsd-tags", []string{}, "Datadog tags, specified as key:value")

	cmdFlags.String("statsd-addr", "", "Statsd address")

	cmdFlags.Bool("enable-prometheus", false, "Enable serving prometheus metrics")

	cmdFlags.Bool("disable-usage-stats", cfg.DisableUsageStats, "Disable sending anonymous usage stats")

	return cmdFlags
}
