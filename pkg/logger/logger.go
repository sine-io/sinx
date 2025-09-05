package logger

import (
	"github.com/sine-io/sinx/pkg/config"

	"go.uber.org/zap"
)

var logger *zap.Logger

func Init() error {
	cfg := config.Get()

	var zapConfig zap.Config

	if cfg.AppEnv == "production" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// 设置日志级别
	switch cfg.LogLevel {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	var err error
	logger, err = zapConfig.Build()
	if err != nil {
		return err
	}

	return nil
}

func Debug(msg string, fields ...interface{}) {
	logger.Sugar().Debugw(msg, fields...)
}

func Info(msg string, fields ...interface{}) {
	logger.Sugar().Infow(msg, fields...)
}

func Warn(msg string, fields ...interface{}) {
	logger.Sugar().Warnw(msg, fields...)
}

func Error(msg string, fields ...interface{}) {
	logger.Sugar().Errorw(msg, fields...)
}

func Fatal(msg string, fields ...interface{}) {
	logger.Sugar().Fatalw(msg, fields...)
}

func GetLogger() *zap.Logger {
	return logger
}

// Sync 刷新缓冲区，确保日志写出
func Sync() {
	if logger != nil {
		_ = logger.Sync() // 忽略常见的 sync 错误（如 Windows 上的 invalid argument）
	}
}
