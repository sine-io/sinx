package log

import (
	"encoding/json"

	"github.com/hashicorp/go-hclog"
	"github.com/rs/zerolog"
)

// HclogHook creates a new hclog.Logger that uses zerolog for output.
// It parses the log message as JSON and adds the fields to the zerolog event.
// The logger is configured to output in JSON format.
// The logger name is set to the provided name.
func HclogHook(name string, logger *zerolog.Logger) hclog.Logger {
	return hclog.New(&hclog.LoggerOptions{
		Name:  name,
		Level: hclog.LevelFromString(zerolog.GlobalLevel().String()),
		Output: logger.Hook(
			zerolog.HookFunc(func(e *zerolog.Event, l zerolog.Level, msg string) {
				var jsonFields map[string]any

				err := json.Unmarshal([]byte(msg), &jsonFields)
				if err != nil {
					e.Err(err)
				} else {
					jsonFields["level"] = jsonFields["@level"]

					// TODO: I don't know how to delete message in zlog.Logger, i will do it later.
					e.Fields(jsonFields)
				}
			}),
		),
		JSONFormat: true,
	})
}
