package flagset

import (
	flag "github.com/spf13/pflag"

	sxconfig "github.com/sine-io/sinx/internal/config"
)

// LogFlagSet creates all of our logging flags.
func LogFlagSet(cfg *sxconfig.Config) *flag.FlagSet {
	cmdFlags := flag.NewFlagSet("logging flagset", flag.ContinueOnError)

	cmdFlags.String("log-level", cfg.LogLevel,
		"Log level (debug|info|warn|error|fatal|panic)")

	return cmdFlags
}
