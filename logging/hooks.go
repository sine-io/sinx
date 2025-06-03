package logging

import (
	"github.com/rs/zerolog"
)

type LogSplitter struct{}

// TODO: Some hooks are not used yet, but they can be useful in the future.
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
