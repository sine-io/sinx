package config

import (
	"fmt"
	"log"
	"os"
	"time"
)

// DefaultBindPort is the default port that dkron will use for Serf communication
const (
	DefaultBindPort      int           = 8946
	DefaultRPCPort       int           = 6868
	DefaultRetryInterval time.Duration = time.Second * 30
)

// DefaultConfig returns a Config struct pointer with sensible
// default settings.
func DefaultConfig() *Config {
	hostname, err := os.Hostname()
	if err != nil {
		log.Panic(err)
	}

	tags := map[string]string{}

	return &Config{
		// ------ configuration for node ------
		NodeName:   hostname,
		Tags:       tags,
		LogLevel:   "info",
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
		DataDir:           "sinx.data",
		ReconcileInterval: 60 * time.Second,
		RaftMultiplier:    1,

		// ------ configuration for cluster ------
		RetryJoinIntervalLAN: DefaultRetryInterval,

		// ------ configuration for observability ------
		MailSubjectPrefix: "[SinX]",
	}
}
