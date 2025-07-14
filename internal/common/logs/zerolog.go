package logs

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Default values for logging configuration
	LogLevel      string = "info"
	LogFilename   string = "sinx.data/sinx.log"
	LogMaxSize    int    = 5
	LogMaxAge     int    = 14
	LogMaxBackups int    = 10
	LogCompress   bool   = true
)

// InitLogger init zerolog.Logger
// Notice, zerolog.Logger is a struct, not a pointer. In Go:
//  1. basic data types and structs: during assignment, the value is copied,
//     so if you want to modify the original value, you need to use a pointer.
//  2. slices, maps, channels, interfaces, functions: during assignment, the reference is copied,
//     so if you want to modify the original value, you can use the value directly.
//  3. pointers: during assignment, the pointer is copied,
//     so if you want to modify the original value, you can use the pointer directly.
//
// In one sentence, we should use a logger pointer to crate agent logger,
//
//	so we can use the initialization which initialized in this function.
//
// Like this:
//
//	agentLogger := &zlog.Logger
//	agentLogger.Hook()
//
// `agentLogger.Hook()` will return a new logger with the same configuration as the global logger, but we can add hooks to it.
func InitLogger() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = time.RFC3339Nano

	LogLevel = strings.ToLower(LogLevel) // normalize the log level to lower case

	parsedLogLevel, _ := zerolog.ParseLevel(LogLevel) // will return NoLevel if invalid, so we ignore the error.
	switch parsedLogLevel {
	case zerolog.NoLevel:
		zerolog.SetGlobalLevel(zerolog.InfoLevel) // NoLevel will be set to InfoLevel.
	case zerolog.FatalLevel, zerolog.PanicLevel:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(parsedLogLevel)
	}

	fileLogger := &lumberjack.Logger{
		Filename:   LogFilename,
		MaxSize:    LogMaxSize,
		MaxBackups: LogMaxBackups,
		MaxAge:     LogMaxAge,
		Compress:   LogCompress,
	}
	writers := zerolog.MultiLevelWriter(zerolog.NewConsoleWriter(), fileLogger)

	// Customize the global logger (the one used by package level methods).
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	zlog.Logger = zerolog.New(writers).
		With().
		Timestamp().
		Caller(). // Add file and line number to log
		Logger()
}
