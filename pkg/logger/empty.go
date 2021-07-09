package logger

import "context"

// Empty useful in testing purposes.
// It doesn't log anything.
type Empty struct{}

func (e Empty) Debug(message string) {
}

func (e Empty) Debugf(format string, args ...interface{}) {
}

func (e Empty) Error(message string) {
}

func (e Empty) Errorf(format string, args ...interface{}) {
}

func (e Empty) Info(message string) {
}

func (e Empty) Infof(format string, args ...interface{}) {
}

func (e Empty) Warn(message string) {
}

func (e Empty) Warnf(format string, args ...interface{}) {
}

func (e Empty) Fatal(message string) {
}

func (e Empty) Fatalf(format string, args ...interface{}) {
}

func (e Empty) Log(msg string) {
}

func (e Empty) With(fields Fields) Logger {
	return e
}

func (e Empty) WithContext(ctx context.Context) Logger {
	return e
}
