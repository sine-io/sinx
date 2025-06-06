package cmd

import (
	"path/filepath"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"

	sxconfig "github.com/sine-io/sinx/internal/config"
)

var (
	cfg = sxconfig.DefaultConfig()

	// Default values for logging configuration
	logLevel      string = "info"
	logFilename   string = "sinx.data/sinx.log"
	logMaxSize    int    = 5
	logMaxAge     int    = 14
	logMaxBackups int    = 10
	logCompress   bool   = true

	rpcAddr string
	ip      string
)

func init() {
	// Initialize the logger
	cobra.OnInitialize(initLogger)

	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", logLevel,
		`Log level (debug, info, warn, error, fatal, panic).
If set, it will override 'log-level' in the config file which initilazed in agent command via viper.
Invalid log level will be set to 'info'.`)
	rootCmd.PersistentFlags().StringVar(&logFilename, "log-filename", logFilename,
		`The file to write logs to. Used by the lumberjack logger.`)
	rootCmd.PersistentFlags().IntVar(&logMaxSize, "log-max-size", logMaxSize,
		`The maximum size in megabytes of the log file before it gets rotated. 
Used by the lumberjack logger.`)
	rootCmd.PersistentFlags().IntVar(&logMaxAge, "log-max-age", logMaxAge,
		`The maximum number of days to retain old log files based on the timestamp encoded in their filename. 
Note that a day is defined as 24 hours and may not exactly correspond to calendar days due to daylight savings, leap seconds, etc. 
Used by the lumberjack logger.`)
	rootCmd.PersistentFlags().IntVar(&logMaxBackups, "log-max-backups", logMaxBackups,
		`The maximum number of old log files to retain, though 'log-max-age' may still cause them to get deleted. 
Used by the lumberjack logger.`)
	rootCmd.PersistentFlags().Bool("log-compress", logCompress,
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

// initLogger init zerolog.Logger
func initLogger() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = time.RFC3339Nano

	parsedLogLevel, err := zerolog.ParseLevel(logLevel)

	if err != nil {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		// zerolog's default level is Debug
		zerolog.SetGlobalLevel(parsedLogLevel)
	}

	fileLogger := &lumberjack.Logger{
		Filename:   logFilename,
		MaxSize:    logMaxSize,
		MaxBackups: logMaxBackups,
		MaxAge:     logMaxAge,
		Compress:   logCompress,
	}
	writers := zerolog.MultiLevelWriter(zerolog.NewConsoleWriter(), fileLogger)

	// Customize the global logger (the one used by package level methods).
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	zlog.Logger = zerolog.New(writers).
		With().
		Timestamp().
		Caller(). // Add file and line number to log
		Logger()
}

// LogSplitter is a zerolog hook that splits logs based on their level.
// It can be used to customize how logs are handled based on their severity.
// TODO: Some hooks are not used yet, but they can be useful in the future.
type LogSplitter struct{}

func (l *LogSplitter) Run(e *zerolog.Event, level zerolog.Level, message string) {
	// switch level {
	// case zerolog.ErrorLevel, zerolog.PanicLevel, zerolog.FatalLevel:
	// 	e.Str("level", "error").Msg(message)
	// case zerolog.WarnLevel, zerolog.DebugLevel, zerolog.InfoLevel, zerolog.TraceLevel:
	// 	e.Str("level", "info").Msg(message)
	// default:
	// 	e.Str("level", "info").Msg(message)
	// }
}

type LogHook struct{}

func (l *LogHook) Run(e *zerolog.Event, level zerolog.Level, message string) {
	e.Str("node", cfg.NodeName)
}
