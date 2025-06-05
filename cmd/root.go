package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sxflagset "github.com/sine-io/sinx/cmd/flagset"
	sxconfig "github.com/sine-io/sinx/internal/config"
)

var (
	cfgFile string
	cfg     = sxconfig.DefaultConfig()

	rpcAddr string
	ip      string

	initErrors []error
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	// add log flags to root command.
	rootCmd.Flags().AddFlagSet(sxflagset.LogFlagSet(cfg))
	_ = viper.BindPFlags(rootCmd.Flags())

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
	if err != nil {             // Handle errors reading the config file
		initErrors = append(initErrors, fmt.Errorf("no valid config found: Applying default values. Error: %s", err))
	}

	if err := viper.Unmarshal(cfg); err != nil {
		initErrors = append(initErrors, fmt.Errorf("error unmarshalling config. Error: %s", err))
	}

	cliTags := viper.GetStringSlice("tag")
	var tags map[string]string

	if len(cliTags) > 0 {
		tags, err = UnmarshalTags(cliTags)
		if err != nil {
			initErrors = append(initErrors, fmt.Errorf("error unmarshalling cli tags. Error: %s", err))
		}
	} else {
		tags = viper.GetStringMapString("tags")
	}
	cfg.Tags = tags

	initLogger()
}

// initLogger init zerolog.Logger
func initLogger() {

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = time.RFC3339Nano

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)

	if err != nil {
		// If the log level is invalid, default to INFO
		logLevel = zerolog.InfoLevel

		initErrors = append(initErrors, fmt.Errorf("invalid log level '%s', defaulting to INFO: %w", logLevel, err))
	}

	// 1. Set the global log level
	// zerolog's default level is Debug
	if logLevel != zerolog.DebugLevel {
		zerolog.SetGlobalLevel(logLevel)
	}

	// 2. Set the output format
	fileLogger := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    5, //
		MaxBackups: 10,
		MaxAge:     14,
		Compress:   true,
	}
	writers := zerolog.MultiLevelWriter(zerolog.NewConsoleWriter(), fileLogger)

	// 3. Customize the global logger (the one used by package level methods).
	zlog.Logger = zerolog.New(writers).
		With().
		Str("node", cfg.NodeName). // Add node information to the logger
		Timestamp().
		Caller(). // Add file and line number to log
		Logger()

	// 4. log the init errors
	for _, err := range initErrors {
		zlog.Error().Err(err).Msg("Logger initialization error")
	}
	initErrors = nil // Clear errors after logging them

	// 5. add hooks
	zlog.Hook(&LogSplitter{})
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
