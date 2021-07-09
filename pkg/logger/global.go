package logger

import (
	"context"
	"io"

	"github.com/rs/zerolog"
)

// std - default global logger.
var std, _ = NewZerologLogger(nil)

func Debug(args string) {
	std.Debug(args)
}

func Debugf(format string, args ...interface{}) {
	std.Debugf(format, args...)
}

func Error(args string) {
	std.Error(args)
}

func Errorf(format string, args ...interface{}) {
	std.Errorf(format, args...)
}
func Info(args string) {
	std.Info(args)
}
func Infof(format string, args ...interface{}) {
	std.Infof(format, args...)
}
func Warn(args string) {
	std.Warn(args)
}
func Warnf(format string, args ...interface{}) {
	std.Warnf(format, args...)
}
func Fatal(args string) {
	std.Fatal(args)
}
func Fatalf(format string, args ...interface{}) {
	std.Fatalf(format, args...)
}
func Log(msg string) {
	std.Log(msg)
}
func With(fields Fields) Logger {
	return std.With(fields)
}

func GetLogger() Logger {
	return std.With(Fields{})
}

func Configure(cfg *Config) (err error) {
	if std, err = NewZerologLogger(cfg); err != nil {
		return err
	}
	return nil
}

func SetLevel(levelStr string) {
	lvl, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	std.Logger = std.Level(lvl)
}

func SetOutput(out io.Writer) {
	std.Logger = std.Output(out)
}

func WithContext(ctx context.Context) Logger {
	return std.WithContext(ctx)
}
