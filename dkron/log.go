package dkron

import (
	"fmt"
	"io"
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

		var output io.Writer = zerolog.ConsoleWriter{
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

		output = zerolog.MultiLevelWriter(os.Stderr, fileLogger)

		formattedLogger = zerolog.New(output).
			Level(logLevel).
			With().Str("node", node). // Add node information to the logger
			Timestamp().
			Logger()
	})

	return formattedLogger

	// formattedLogger := logrus.New()
	// formattedLogger.Formatter = &logrus.TextFormatter{FullTimestamp: true}

	// level, err := logrus.ParseLevel(logLevel)
	// if err != nil {
	// 	logrus.WithError(err).Error("Error parsing log level, using: info")
	// 	level = logrus.InfoLevel
	// }

	// formattedLogger.Level = level
	// log := logrus.NewEntry(formattedLogger).WithField("node", node)

	// ginOnce.Do(func() {
	// 	if level == logrus.DebugLevel {
	// 		gin.DefaultWriter = log.Writer()
	// 		gin.SetMode(gin.DebugMode)
	// 	} else {
	// 		gin.DefaultWriter = io.Discard
	// 		gin.SetMode(gin.ReleaseMode)
	// 	}
	// })

	// return log
}
