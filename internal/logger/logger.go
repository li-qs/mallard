package logger

import (
	"go.uber.org/zap"
)

var l *zap.Logger

// level: "debug"-development, others-production
func Init(level string) error {
	var err error

	if level == "debug" {
		l, err = zap.NewDevelopment()
	} else {
		l, err = zap.NewProduction()
	}

	return err
}

func GetLogger() *zap.Logger {
	if l == nil {
		panic("logger: Init must be called before GetLogger")
	}
	return l
}

func Sync() error {
	if l == nil {
		return nil
	}
	return l.Sync()
}

func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}
