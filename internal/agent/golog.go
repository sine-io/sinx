package agent

import (
	golog "log"
	"regexp"
	"strings"

	"github.com/rs/zerolog"
)

// customGologWithZerolog creates a new golog.Logger that uses zerolog for output.
// It parses the log message to extract the log level from the message format used by go original log.
// TODO: We can't controll the Go log's level, need to enhance this in the future.
func customGologWithZerolog(logger zerolog.Logger) *golog.Logger {
	return golog.New(
		logger.Hook(
			zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
				/*
					message e.g: "[INFO] serf: EventMemberJoin: node1 192.168.0.103"
				*/
				// we can't set go original log format to json, so we use regex to extract log level
				tmpStr := regexp.MustCompile(`\[(\w+)\]`).FindStringSubmatch(msg)
				e.Str("level", strings.ToLower(tmpStr[1]))
			}),
		),
		"", 0,
	)
}
