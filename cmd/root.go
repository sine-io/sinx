package cmd

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sxconfig "github.com/sine-io/sinx/internal/config"
)

var (
	cfgFile string
	cfg     = sxconfig.DefaultConfig()

	rpcAddr string
	ip      string
)

func init() {
	// Initialize the logger
	cobra.OnInitialize(initLogger)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().StringVar(&cfg.LogLevel, "log-level", "info", "log level (debug, info, warn, error, fatal, panic)")

	// cobra.OnFinalize()
}

// rootCmd represents the dkron command
var rootCmd = &cobra.Command{
	Use:   "dkron",
	Short: "Open source distributed job scheduling system",
	Long: `Dkron is a system service that runs scheduled jobs at given intervals or times,
just like the cron unix service but distributed in several machines in a cluster.
If a machine fails (the leader), a follower will take over and keep running the scheduled jobs without human intervention.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("dkron")        // name of config file (without extension)
		viper.AddConfigPath("/etc/dkron")   // call multiple times to add many search paths
		viper.AddConfigPath("$HOME/.dkron") // call multiple times to add many search paths
		viper.AddConfigPath("./config")     // call multiple times to add many search paths
	}

	viper.SetEnvPrefix("dkron")
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv() // read in environment variables that match

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		zlog.Info().Err(err).Msg("No valid config found: Applying default values.")
	}

	if err := viper.Unmarshal(cfg); err != nil {
		zlog.Error().Err(err).Msg("Error unmarshalling config")
	}

	cliTags := viper.GetStringSlice("tag")
	var tags map[string]string

	if len(cliTags) > 0 {
		tags, err = UnmarshalTags(cliTags)
		if err != nil {
			zlog.Error().Err(err).Msg("Error unmarshalling cli tags")
		}
	} else {
		tags = viper.GetStringMapString("tags")
	}
	cfg.Tags = tags
}

// initLogger init zerolog.Logger
func initLogger() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = time.RFC3339Nano

	fileLogger := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    5, //
		MaxBackups: 10,
		MaxAge:     14,
		Compress:   true,
	}
	writers := zerolog.MultiLevelWriter(zerolog.NewConsoleWriter(), fileLogger)

	// Customize the global logger (the one used by package level methods).
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	zlog.Logger = zerolog.New(writers).
		With().
		Str("node", cfg.NodeName). // Add node information to the logger
		Timestamp().
		Caller(). // Add file and line number to log
		Logger()
}

func setupLogLevel() {
	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)

	if err != nil {
		// If the log level is invalid, default to INFO
		zlog.Info().Err(err).Msgf("invalid log level '%s', defaulting to INFO", logLevel)

		logLevel = zerolog.InfoLevel
	}

	// zerolog's default level is Debug
	if logLevel != zerolog.DebugLevel {
		zerolog.SetGlobalLevel(logLevel)
	}
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
