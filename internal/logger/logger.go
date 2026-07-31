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
	return l
}

func Sync() error {
	return l.Sync()
}

func Debug(msg string, fields ...zap.Field) {
	l.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	l.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	l.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	l.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	l.Fatal(msg, fields...)
}
