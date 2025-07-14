package decorator

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
)

type commandLoggingDecorator[C any] struct {
	base   CommandHandler[C]
	logger *zerolog.Logger
}

func (d commandLoggingDecorator[C]) Handle(ctx context.Context, cmd C) (err error) {
	handlerType := generateActionName(cmd)

	logger := d.logger.Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
		e.Str("command", handlerType)
		e.Str("command_body", fmt.Sprintf("%#v", cmd))
	}))

	logger.Debug().Msg("Executing command")

	defer func() {
		if err == nil {
			logger.Info().Msg("Command executed successfully")
		} else {
			logger.Error().Err(err).Msg("Failed to execute command")
		}
	}()

	return d.base.Handle(ctx, cmd)
}

type queryLoggingDecorator[C any, R any] struct {
	base   QueryHandler[C, R]
	logger *zerolog.Logger
}

func (d queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	logger := d.logger.Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
		e.Str("query", generateActionName(cmd))
		e.Str("query_body", fmt.Sprintf("%#v", cmd))
	}))

	logger.Debug().Msg("Executing query")

	defer func() {
		if err == nil {
			logger.Info().Msg("Query executed successfully")
		} else {
			logger.Error().Err(err).Msg("Failed to execute query")
		}
	}()

	return d.base.Handle(ctx, cmd)
}
