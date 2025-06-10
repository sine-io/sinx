package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

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
