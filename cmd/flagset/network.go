package flagset

import (
	flag "github.com/spf13/pflag"

	sxcfg "github.com/sine-io/sinx/internal/config"
)

// NetworkFlagSet creates all of our network flags.
func NetworkFlagSet(cfg *sxcfg.Config) *flag.FlagSet {
	cmdFlags := flag.NewFlagSet("network flagset", flag.ContinueOnError)

	cmdFlags.String("bind-addr", cfg.BindAddr,
		`Specifies which address the agent should bind to for network services, 
including the internal gossip protocol and RPC mechanism. This should be 
specified in IP format, and can be used to easily bind all network services 
to the same address. The value supports go-sockaddr/template format.
`)

	cmdFlags.String("http-addr", cfg.HTTPAddr,
		`Address to bind the UI web server to. Only used when server. The value 
supports go-sockaddr/template format.`)

	cmdFlags.String("advertise-addr", "",
		`Address used to advertise to other nodes in the cluster. By default,
the bind address is advertised. The value supports 
go-sockaddr/template format.`)

	cmdFlags.Int("rpc-port", cfg.RPCPort,
		`RPC Port used to communicate with clients. Only used when server. 
The RPC IP Address will be the same as the bind address.`)

	cmdFlags.Int("advertise-rpc-port", 0,
		"Use the value of rpc-port by default")

	cmdFlags.String("serf-reconnect-timeout", cfg.SerfReconnectTimeout,
		`This is the amount of time to attempt to reconnect to a failed node before 
giving up and considering it completely gone. In Kubernetes, you might need 
this to about 5s, because there is no reason to try reconnects for default 
24h value. Also Raft behaves oddly if node is not reaped and returned with 
same ID, but different IP.
Format there: https://golang.org/pkg/time/#ParseDuration`)

	return cmdFlags
}
