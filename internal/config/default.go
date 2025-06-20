package config

import (
	"fmt"
	"os"
	"time"

	zlog "github.com/rs/zerolog/log"
)

// DefaultBindPort is the default port that sinx will use for Serf communication
const (
	DefaultDataDir       string        = "sinx.data"
	DefaultBindPort      int           = 8946
	DefaultRPCPort       int           = 6868
	DefaultRetryInterval time.Duration = time.Second * 30
)

// DefaultConfig returns a Config struct pointer with sensible
// default settings.
func DefaultConfig() *Config {
	hostname, err := os.Hostname()
	if err != nil {
		zlog.Panic().Err(err).Send() // TODO: will call os.Exit(1)? if called, hostname will not be 'anonymous'
	} else {
		hostname = "anonymous"
	}

	tags := map[string]string{}

	return &Config{
		// ------ configuration for node ------
		NodeName:   hostname,
		Tags:       tags,
		Datacenter: "dc1",
		Region:     "global",
		UI:         true,
		Profile:    "lan",

		// ------ configuration for network ------
		BindAddr:             fmt.Sprintf("{{ GetPrivateIP }}:%d", DefaultBindPort),
		HTTPAddr:             ":8080",
		RPCPort:              DefaultRPCPort,
		SerfReconnectTimeout: "24h",

		// ------ configuration for storage ------
		ReconcileInterval: 60 * time.Second,
		RaftMultiplier:    1,
		DataDir:           DefaultDataDir,

		// ------ configuration for cluster ------
		RetryJoinIntervalLAN: DefaultRetryInterval,

		// ------ configuration for observability ------
		MailSubjectPrefix: "[SinX]",
		DisableUsageStats: true,
	}
}
