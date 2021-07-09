package logger

import (
	"context"
	"io"
)

type ctxLoggerFields string
type ctxLoggerKey uint8

// DefaultLevel represents default logger level
const DefaultLevel = "debug"
const loggerFieldsKey ctxLoggerFields = "fields"
const loggerKey ctxLoggerKey = 1

type Config struct {
	Raw           bool
	Output        io.Writer
	Level         string
	OmitTimestamp bool
}

// Logger provides a leveled-zerolog interface.
type Logger interface {
	// Leveled methods, from logrus
	Debug(args string)
	Debugf(format string, args ...interface{})

	Error(args string)
	Errorf(format string, args ...interface{})

	Info(args string)
	Infof(format string, args ...interface{})

	Warn(args string)
	Warnf(format string, args ...interface{})

	Fatal(args string)
	Fatalf(format string, args ...interface{})

	Log(msg string)

	With(fields Fields) Logger

	WithContext(ctx context.Context) Logger
}

// NewContext provides fields to log inside given context and returns new
// value context.
func NewContext(ctx context.Context, fields Fields) context.Context {
	fieldsInside := GetFieldsFromContext(ctx)
	// merge fields with existing ones, to prevent them from being lost
	return context.WithValue(ctx, loggerFieldsKey, fields.Merge(fieldsInside))
}

// NewContextWithLogger returns new context, with added logger instance as a value.
func NewContextWithLogger(ctx context.Context, log Logger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

// GetLoggerFromContext returns logger instance from given context.
// If logger hasn't been set to context yet, then it returns new instance of Empty logger.
func GetLoggerFromContext(ctx context.Context) Logger {
	l, ok := ctx.Value(loggerKey).(Logger)
	if !ok || l == nil {
		// fallback to prevent nil pointer errors
		// TODO: (i.malyuk) do we really need this fallback
		return &Empty{}
	}

	return l
}

// GetFieldsFromContext return zerolog fields from context.
func GetFieldsFromContext(ctx context.Context) (ret Fields) {
	fields, ok := ctx.Value(loggerFieldsKey).(Fields)
	if !ok {
		return
	}

	return fields
}

// Fields map representing structure log.
type Fields map[string]interface{}

// Merge merges fields with given.
func (f Fields) Merge(fields Fields) Fields {
	for k, v := range fields {
		f[k] = v
	}

	return f
}
