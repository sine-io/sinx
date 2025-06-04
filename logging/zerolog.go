package logging

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/sine-io/sinx/internal/config"
)

var (
	once       sync.Once
	initErrors []error
	L          zerolog.Logger
)

// GetLogger creates the logger instance
func GetLogger(cfg *config.Config) zerolog.Logger {

	once.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel, err := zerolog.ParseLevel(cfg.LogLevel)

		if err != nil {
			// If the log level is invalid, default to INFO
			logLevel = zerolog.InfoLevel

			initErrors = append(initErrors, fmt.Errorf("invalid log level '%s', defaulting to INFO: %w", logLevel, err))
		}

		consoleLogger := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		fileLogger := &lumberjack.Logger{
			Filename:   "logs/app.log",
			MaxSize:    5, //
			MaxBackups: 10,
			MaxAge:     14,
			Compress:   true,
		}

		writer := zerolog.MultiLevelWriter(consoleLogger, fileLogger)

		L = zerolog.New(writer).
			Level(logLevel).
			With().Str("node", cfg.NodeName). // Add node information to the logger
			Timestamp().
			Logger()

		for _, err := range initErrors {
			L.Error().Err(err).Msg("Logger initialization error")
		}
		initErrors = nil // Clear errors after logging them

		L.Hook(&LogSplitter{})
	})

	return L
}
