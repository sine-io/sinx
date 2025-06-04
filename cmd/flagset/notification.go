package flagset

import (
	flag "github.com/spf13/pflag"

	"github.com/sine-io/sinx/internal/config"
)

// NotificationFlagSet creates all of our notification flags.
func NotificationFlagSet(cfg *config.Config) *flag.FlagSet {
	cmdFlags := flag.NewFlagSet("notification flagset", flag.ContinueOnError)

	cmdFlags.String("mail-host", "", "Mail server host address to use for notifications")
	cmdFlags.Uint16("mail-port", 0, "Mail server port")
	cmdFlags.String("mail-username", "", "Mail server username used for authentication")
	cmdFlags.String("mail-password", "", "Mail server password to use")
	cmdFlags.String("mail-from", "", "From email address to use")
	cmdFlags.String("mail-payload", "", "Notification mail payload")
	cmdFlags.String("mail-subject-prefix", cfg.MailSubjectPrefix, "Notification mail subject prefix")

	cmdFlags.String("pre-webhook-endpoint", "", "Pre-webhook endpoint to call for notifications")
	cmdFlags.String("pre-webhook-payload", "", "Body of the POST request to send on pre-webhook call")
	cmdFlags.StringSlice("pre-webhook-headers", []string{}, "Headers to use when calling the pre-webhook. Can be specified multiple times")

	cmdFlags.String("webhook-endpoint", "", "Webhook endpoint to call for notifications")
	cmdFlags.String("webhook-payload", "", "Body of the POST request to send on webhook call")
	cmdFlags.StringSlice("webhook-headers", []string{}, "Headers to use when calling the webhook URL. Can be specified multiple times")

	cmdFlags.String("cronitor-endpoint", "", "Cronitor endpoint to call for notifications")

	return cmdFlags
}
