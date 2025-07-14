package logs

import (
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

func LogCommandExecution(commandName string, cmd any, err error) {
	// TODO: I don't know whether it is useful or not. noqa.
	log := zlog.Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
		e.Any("cmd", cmd)
	}))

	log = zlog.With().Any("cmd", cmd).Logger()

	if err == nil {
		log.Info().Msg(commandName + " command succeeded")
	} else {
		log.Error().Err(err).Msg(commandName + " command failed")
	}
}
