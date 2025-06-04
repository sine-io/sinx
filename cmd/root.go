package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sine-io/sinx/cmd/flagset"
	sconfig "github.com/sine-io/sinx/internal/config"
	slogging "github.com/sine-io/sinx/logging"
)

var (
	cfgFile    string
	rpcAddr    string
	ip         string
	initErrors []error

	GlobalCfg = sconfig.DefaultConfig() // GlobalCfg holds the global configuration for the application
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	// add log flags to root command.
	rootCmd.Flags().AddFlagSet(flagset.LogFlagSet(GlobalCfg))
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

	if err := viper.Unmarshal(GlobalCfg); err != nil {
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
	GlobalCfg.Tags = tags

	// logging.L will be initialized with the global configuration
	slogging.GetLogger(GlobalCfg)

	if len(initErrors) > 0 {
		for _, err := range initErrors {
			slogging.L.Error().Err(err).Msg("Initialization error")
		}
	} else {
		slogging.L.Info().Msg("Configuration loaded successfully")
	}
}
