package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type zerologLogger struct {
	zerolog.Logger
}

func (z *zerologLogger) Debug(message string) {
	z.Logger.Debug().Msg(message)
}

func (z *zerologLogger) Debugf(format string, args ...interface{}) {
	z.Logger.Debug().Msgf(format, args...)
}

func (z *zerologLogger) Error(message string) {
	z.Logger.Error().Msg(message)
}

func (z *zerologLogger) Errorf(format string, args ...interface{}) {
	z.Logger.Error().Msgf(format, args...)
}

func (z *zerologLogger) Info(message string) {
	z.Logger.Info().Msg(message)
}

func (z *zerologLogger) Infof(format string, args ...interface{}) {
	z.Logger.Info().Msgf(format, args...)
}

func (z zerologLogger) Warn(message string) {
	z.Logger.Warn().Msg(message)
}

func (z *zerologLogger) Warnf(format string, args ...interface{}) {
	z.Logger.Warn().Msgf(format, args...)
}

func (z *zerologLogger) Fatal(message string) {
	z.Logger.Fatal().Msg(message)
}

func (z *zerologLogger) Fatalf(format string, args ...interface{}) {
	z.Logger.Fatal().Msgf(format, args...)
}

func (z *zerologLogger) Log(msg string) {
	z.Logger.Log().Msg(msg)
}

func (z *zerologLogger) With(fields Fields) Logger {
	newLogger := *z
	newLogger.Logger = newLogger.Logger.With().Fields(fields).Logger()

	return &newLogger
}

func (z *zerologLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return z
	}

	if fields, ok := ctx.Value(loggerFieldsKey).(Fields); ok {
		return z.With(fields)
	}

	return z
}

func NewZerologLogger(cfg *Config) (*zerologLogger, error) {
	if cfg == nil {
		cfg = &Config{
			Level:  DefaultLevel,
			Output: os.Stderr,
		}
	}

	if cfg.Raw {
		cfg.Output = zerolog.ConsoleWriter{TimeFormat: time.RFC3339, Out: cfg.Output}
	}

	if cfg.Level == "" {
		cfg.Level = DefaultLevel
	}

	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse level \"%s\": %s", level, err)
	}

	log := zerolog.New(cfg.Output).Level(level)
	if !cfg.OmitTimestamp {
		log = log.With().Timestamp().Logger()
	}

	return &zerologLogger{
		Logger: log,
	}, nil
}
