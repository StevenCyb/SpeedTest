package utils

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger create a new zap logger.
func InitLogger(level zap.AtomicLevel) (*zap.Logger, error) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = level
	zapConfig.Encoding = "console"
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	zapConfig.EncoderConfig.TimeKey = "time"
	zapConfig.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.UTC().Format("02.01.2006-15:04:05"))
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("error building logger: %w", err)
	}

	return logger, nil
}

// LevelFromString transform string to zap log level:
/*
 * debug|DEBUG
 * ino|INFO
 * warn|WARN
 * error|ERROR
 * dpanic|DPANIC
 * panic|PANIC
 * fatal|FATAL
 *.
 */
func LevelFromString(logLevel string) zapcore.Level {
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		return zap.DebugLevel
	case "INFO", "":
		return zap.InfoLevel
	case "WARN":
		return zap.WarnLevel
	case "ERROR":
		return zap.ErrorLevel
	case "DPANIC":
		return zap.DPanicLevel
	case "PANIC":
		return zap.PanicLevel
	case "FATAL":
		return zap.FatalLevel
	}

	return zap.InfoLevel
}

// SelectiveLog execute the log type according to the given logLevel.
//
//nolint:exhaustive
func SelectiveLog(logger *zap.SugaredLogger, logLevel zapcore.Level, logs string) {
	switch logLevel {
	case zap.DebugLevel:
		logger.Debug(logs)
	case zap.InfoLevel:
		logger.Info(logs)
	case zap.WarnLevel:
		logger.Warn(logs)
	case zap.ErrorLevel:
		logger.Error(logs)
	case zap.DPanicLevel:
		logger.DPanic(logs)
	case zap.PanicLevel:
		logger.Panic(logs)
	case zap.FatalLevel:
		logger.Fatal(logs)
	}
}
