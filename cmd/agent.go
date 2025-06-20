package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sxflagset "github.com/sine-io/sinx/cmd/flagset"
	sxagent "github.com/sine-io/sinx/internal/agent"
	sxconfig "github.com/sine-io/sinx/internal/config"
	sxui "github.com/sine-io/sinx/internal/ui"
	sxplugin "github.com/sine-io/sinx/plugin"
)

var (
	cfgFile string
	cfg     = sxconfig.DefaultConfig()

	ShutdownCh chan struct{}
	agent      *sxagent.Agent

	logger = zerolog.New(zerolog.NewConsoleWriter())
)

const (
	// gracefulTimeout controls how long we wait before forcefully terminating
	gracefulTimeout = 3 * time.Hour
)

func init() {
	rootCmd.AddCommand(agentCmd)
	agentCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file path")

	agentCmd.Flags().AddFlagSet(sxflagset.NodeFlagSet(cfg))
	agentCmd.Flags().AddFlagSet(sxflagset.NetworkFlagSet(cfg))
	agentCmd.Flags().AddFlagSet(sxflagset.ClusterFlagSet(cfg))
	agentCmd.Flags().AddFlagSet(sxflagset.StorageFlagSet(cfg))
	agentCmd.Flags().AddFlagSet(sxflagset.NotificationFlagSet(cfg))
	agentCmd.Flags().AddFlagSet(sxflagset.ObservabilityFlagSet(cfg))

	_ = viper.BindPFlags(agentCmd.Flags())
}

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Start a sinx agent",
	Long: `Start a sinx agent that schedules jobs, listens for executions and runs executors.
It also runs a web UI.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig()
	},
	// Run will execute the main functions of the agent command.
	// This includes the main eventloop and starting the server if enabled.
	//
	// The returned value is the exit code.
	// protoc -I proto/ proto/executor.proto --go_out=plugins=grpc:sinx/
	RunE: func(cmd *cobra.Command, args []string) error {
		return agentRun(args...)
	},
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("sinx")        // name of config file (without extension)
		viper.AddConfigPath("/etc/sinx")   // call multiple times to add many search paths
		viper.AddConfigPath("$HOME/.sinx") // call multiple times to add many search paths
		viper.AddConfigPath("./config")    // call multiple times to add many search paths
	}

	viper.SetEnvPrefix("sinx")
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv() // read in environment variables that match

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		return fmt.Errorf("config: Error reading config file: %s", err.Error())
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("config: Error unmarshalling config: %s", err.Error())
	}

	cliTags := viper.GetStringSlice("tag")
	var tags map[string]string

	if len(cliTags) > 0 {
		tags, err = UnmarshalTags(cliTags)
		if err != nil {
			return fmt.Errorf("config: Error unmarshalling cli tags: %s", err.Error())
		}
	} else {
		tags = viper.GetStringMapString("tags")
	}
	cfg.Tags = tags

	return nil
}

func agentRun(args ...string) error {
	// 1. init agent with config and logger.
	agent = sxagent.NewAgent(cfg).WithLogger(&zlog.Logger)

	logger = agent.Logger().Hook()
	// This log statement helps avoid compiler warnings about unused parameters
	// as 'args' is not used elsewhere in the function
	logger.Debug().Msgf("agentRun called with args: %v", args)

	// 2. set agent plugins
	p := sxplugin.NewPlugins().WithLogger(&logger)
	if err := p.DiscoverPlugins(); err != nil {
		logger.Fatal().Err(err).Send()
	}

	plugins := sxagent.PluginRegistry{
		Processors: p.Processors,
		Executors:  p.Executors,
	}
	agent.WithPlugins(plugins)

	// 3. set agent transport
	agent.HTTPTransport = sxui.NewHTTPTransport(agent).WithLogger(agent.Logger())

	// 4. start the agent
	// TODO: init all the components of the agent in the StartAgent method.
	if err := agent.StartAgent(); err != nil {
		return err
	}

	// 5. handle signals
	exit := handleSignals()
	if exit != 0 {
		return fmt.Errorf("exit status: %d", exit)
	}

	return nil
}

// handleSignals blocks until we get an exit-causing signal
func handleSignals() int {
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

WAIT:
	// Wait for a signal
	var sig os.Signal
	select {
	case s := <-signalCh:
		sig = s
	case err := <-agent.RetryJoinCh():
		logger.Error().Err(err).Msg("agent: Retry join failed")
		return 1
	case <-ShutdownCh:
		sig = os.Interrupt
	}
	logger.Info().Msgf("Caught signal: %v", sig)

	// Check if this is a SIGHUP
	if sig == syscall.SIGHUP {
		handleReload()
		goto WAIT
	}

	// Fail fast if not doing a graceful leave
	if sig != syscall.SIGTERM && sig != os.Interrupt {
		return 1
	}

	// Attempt a graceful leave
	logger.Info().Msg("agent: Gracefully shutting down agent...")
	go func() {
		if err := agent.StopAgent(); err != nil {
			logger.Error().Err(err).Msg("unable to stop agent")
			return
		}
	}()

	gracefulCh := make(chan struct{})

	for {
		logger.Info().Msg("Waiting for jobs to finish...")
		if agent.GetRunningJobs() < 1 {
			logger.Info().Msg("No jobs left. Exiting.")
			break
		}
		time.Sleep(1 * time.Second)
	}

	plugin.CleanupClients()

	close(gracefulCh)

	// Wait for leave or another signal
	select {
	case <-signalCh:
		return 1
	case <-time.After(gracefulTimeout):
		return 1
	case <-gracefulCh:
		return 0
	}
}

// handleReload is invoked when we should reload our configs, e.g. SIGHUP
func handleReload() {
	logger.Info().Msg("Reloading configuration...")
	initConfig()
	//Config reloading will also reload Notification settings
	agent.UpdateTags(cfg.Tags)
}

// UnmarshalTags is a utility function which takes a slice of strings in
// key=value format and returns them as a tag mapping.
func UnmarshalTags(tags []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, tag := range tags {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) != 2 || len(parts[0]) == 0 {
			return nil, fmt.Errorf("invalid tag: '%s'", tag)
		}
		result[parts[0]] = parts[1]
	}
	return result, nil
}
