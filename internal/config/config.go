package config

import (
	"time"
)

// Config stores all configuration options.
type Config struct {
	// ------ configuration for node ------
	// NodeName is the name we register as. Defaults to hostname.
	NodeName string `mapstructure:"node-name"`

	// Tags are used to attach key/value metadata to a node.
	Tags map[string]string `mapstructure:"tags"`

	// Datacenter is the datacenter this server belongs to.
	Datacenter string

	// Region is the region this server belongs to.
	Region string

	// Server enables this node to work as a server.
	Server bool

	// UI enable the web UI on this node. The node must be server.
	UI bool

	// DevMode is used for development purposes only and limits the
	// use of persistence or state.
	DevMode bool

	// Profile is used to select a timing profile for Serf. The supported choices
	// are "wan", "lan", and "local". The default is "lan"
	Profile string
	// ------ end ------

	// ------ configuration for log ------
	// LogLevel is the log verbosity level used.
	// It can be (debug|info|warn|error|fatal|panic).
	LogLevel string `mapstructure:"log-level"`
	// ------ end ------

	// ------ configuration for network ------
	// BindAddr is the address on which all of services will
	// be bound. If not specified, this defaults to the first private ip address.
	BindAddr string `mapstructure:"bind-addr"`

	// HTTPAddr is the address on the UI web server will
	// be bound. If not specified, this defaults to all interfaces.
	HTTPAddr string `mapstructure:"http-addr"`

	// AdvertiseAddr is the address that the Serf and gRPC layer will advertise to
	// other members of the cluster. Can be used for basic NAT traversal
	// where both the internal ip:port and external ip:port are known.
	AdvertiseAddr string `mapstructure:"advertise-addr"`

	// RPCPort is the gRPC port. This should be reachable
	// by the other servers and clients.
	RPCPort int `mapstructure:"rpc-port"`

	// AdvertiseRPCPort is the gRPC port advertised to clients. This should be reachable
	// by the other servers and clients.
	AdvertiseRPCPort int `mapstructure:"advertise-rpc-port"`

	// SerfReconnectTimeout is the amount of time to attempt to reconnect to a failed node
	// before giving up and considering it completely gone
	SerfReconnectTimeout string `mapstructure:"serf-reconnect-timeout"`
	// ------ end ------

	// ------ configuration for cluster ------
	// EncryptKey is the secret key to use for encrypting communication
	// traffic for Serf. The secret key must be exactly 32-bytes, base64
	// encoded. The easiest way to do this on Unix machines is this command:
	// "head -c32 /dev/urandom | base64" or use "sinx keygen". If this is
	// not specified, the traffic will not be encrypted.
	EncryptKey string `mapstructure:"encrypt"`

	// StartJoin is a list of addresses to attempt to join when the
	// agent starts. If Serf is unable to communicate with any of these
	// addresses, then the agent will error and exit.
	StartJoin []string `mapstructure:"join"`

	// RetryJoinLAN is a list of addresses to attempt to join when the
	// agent starts. Serf will continue to retry the join until it
	// succeeds or RetryMaxAttempts is reached.
	RetryJoinLAN []string `mapstructure:"retry-join"`

	// RetryMaxAttemptsLAN is used to limit the maximum attempts made
	// by RetryJoin to reach other nodes. If this is 0, then no limit
	// is imposed, and Serf will continue to try forever. Defaults to 0.
	RetryJoinMaxAttemptsLAN int `mapstructure:"retry-max"`

	// RetryIntervalLAN is the string retry interval. This interval
	// controls how often we retry the join for RetryJoin. This defaults
	// to 30 seconds.
	RetryJoinIntervalLAN time.Duration `mapstructure:"retry-interval"`

	// Bootstrap mode is used to bring up the first server.  It is
	// required so that it can elect a leader without any other nodes
	// being present
	Bootstrap bool

	// BootstrapExpect tries to automatically bootstrap the cluster,
	// by withholding peers until enough servers join.
	BootstrapExpect int `mapstructure:"bootstrap-expect"`
	// ------ end ------

	// ------ configuration for storage ------
	// DataDir is the directory to store our state in
	DataDir string `mapstructure:"data-dir"`

	// ReconcileInterval controls how often we reconcile the strongly
	// consistent store with the Serf info. This is used to handle nodes
	// that are force removed, as well as intermittent unavailability during
	// leader election.
	ReconcileInterval time.Duration

	// RaftMultiplier An integer multiplier used by servers to scale key
	// Raft timing parameters.
	RaftMultiplier int `mapstructure:"raft-multiplier"`
	// ------ end ------

	// ------ configuration for notifications ------
	// MailHost is the SMTP server host to use for email notifications.
	MailHost string `mapstructure:"mail-host"`

	// MailPort is the SMTP server port to use for email notifications.
	MailPort uint16 `mapstructure:"mail-port"`

	// MailUsername is the SMTP server username to use for email notifications.
	MailUsername string `mapstructure:"mail-username"`

	// MailPassword is the SMTP server password to use for email notifications.
	MailPassword string `mapstructure:"mail-password"`

	// MailFrom is the email sender to use for email notifications.
	MailFrom string `mapstructure:"mail-from"`

	// MailPayload is the email template body to use for email notifications.
	MailPayload string `mapstructure:"mail-payload"`

	// MailSubjectPrefix is the email subject prefix string to use for email notifications.
	MailSubjectPrefix string `mapstructure:"mail-subject-prefix"`

	// PreWebhookURL is the endpoint to call for notifications.
	PreWebhookEndpoint string `mapstructure:"pre-webhook-endpoint"`

	// PreWebhookPayload is the body template of the request for notifications.
	PreWebhookPayload string `mapstructure:"pre-webhook-payload"`

	// PreWebhookHeaders are the headers to use when calling the webhook for notifications.
	PreWebhookHeaders []string `mapstructure:"pre-webhook-headers"`

	// WebhookEndpoint is the URL to call for notifications.
	WebhookEndpoint string `mapstructure:"webhook-endpoint"`

	// WebhookPayload is the body template of the request for notifications.
	WebhookPayload string `mapstructure:"webhook-payload"`

	// WebhookHeaders are the headers to use when calling the webhook for notifications.
	WebhookHeaders []string `mapstructure:"webhook-headers"`

	// CronitorEndpoint is the endpoint to call for cronitor notifications.
	CronitorEndpoint string `mapstructure:"cronitor-endpoint"`
	// ------ end ------

	// ------ configuration for observability ------
	// DogStatsdAddr is the address of a dogstatsd instance. If provided,
	// metrics will be sent to that instance.
	DogStatsdAddr string `mapstructure:"dog-statsd-addr"`

	// DogStatsdTags are the global tags that should be sent with each packet to dogstatsd
	// It is a list of strings, where each string looks like "my_tag_name:my_tag_value".
	DogStatsdTags []string `mapstructure:"dog-statsd-tags"`

	// StatsdAddr is the statsd standard server to be used for sending metrics.
	StatsdAddr string `mapstructure:"statsd-addr"`

	// EnablePrometheus enables serving of prometheus metrics at /metrics
	EnablePrometheus bool `mapstructure:"enable-prometheus"`

	// DisableUsageStats disable sending anonymous usage stats
	DisableUsageStats bool `mapstructure:"disable-usage-stats"`
	// ------ end ------
}
