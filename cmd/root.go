package cmd

import (
	"github.com/spf13/cobra"

	sxlog "github.com/sine-io/sinx/log"
)

var (
	rpcAddr string
	ip      string
)

func init() {
	// Initialize the logger
	cobra.OnInitialize(sxlog.InitLogger)

	rootCmd.PersistentFlags().StringVar(&sxlog.LogLevel, "log-level", sxlog.LogLevel,
		`Log level (trace, debug, info, warn, error, fatal, panic, disabled), same to zerolog.
It's case insensitive, so you can use 'DEBUG', 'Info', etc.
Level 'disabled' will disable all logging.
Invalid log level will be set to 'info'.`)
	rootCmd.PersistentFlags().StringVar(&sxlog.LogFilename, "log-filename", sxlog.LogFilename,
		`The file to write logs to. Used by the lumberjack logger.`)
	rootCmd.PersistentFlags().IntVar(&sxlog.LogMaxSize, "log-max-size", sxlog.LogMaxSize,
		`The maximum size in megabytes of the log file before it gets rotated. 
Used by the lumberjack logger.`)
	rootCmd.PersistentFlags().IntVar(&sxlog.LogMaxAge, "log-max-age", sxlog.LogMaxAge,
		`The maximum number of days to retain old log files based on the timestamp encoded in their filename. 
Note that a day is defined as 24 hours and may not exactly correspond to calendar days due to daylight savings, leap seconds, etc. 
Used by the lumberjack logger.`)
	rootCmd.PersistentFlags().IntVar(&sxlog.LogMaxBackups, "log-max-backups", sxlog.LogMaxBackups,
		`The maximum number of old log files to retain, though 'log-max-age' may still cause them to get deleted. 
Used by the lumberjack logger.`)
	rootCmd.PersistentFlags().Bool("log-compress", sxlog.LogCompress,
		`Compress the rotated log files via gzip. Used by the lumberjack logger.`)
}

// rootCmd represents the sinx command
var rootCmd = &cobra.Command{
	Use:   "sinx",
	Short: "Open source distributed job scheduling system",
	Long: `SinX is a system service that runs scheduled jobs at given intervals or times,
just like the cron unix service but distributed in several machines in a cluster.
If a machine fails (the leader), a follower will take over and keep running the scheduled jobs without human intervention.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
