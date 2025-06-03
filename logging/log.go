package logging

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	once            sync.Once
	initErrors      []error
	formattedLogger zerolog.Logger
)

// InitLogger creates the logger instance
func InitLogger(logLevel string, node string) zerolog.Logger {

	once.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel, err := zerolog.ParseLevel(logLevel)

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

		formattedLogger = zerolog.New(writer).
			Level(logLevel).
			With().Str("node", node). // Add node information to the logger
			Timestamp().
			Logger()

		for _, err := range initErrors {
			formattedLogger.Error().Err(err).Msg("Logger initialization error")
		}
		initErrors = nil // Clear errors after logging them
	})

	return formattedLogger
}
