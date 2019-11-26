package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Field log field
type Field = zap.Field

// Option log Option
type Option = zap.Option

// Logger logger
type Logger = zap.Logger

// Level log level
type Level = zapcore.Level

// all log level
const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// var _log, _ = zap.NewDevelopment()

var _log, _ = zap.NewProduction()

// Init initializes logger
func Init(l *Logger) {
	_log = l
}

// Int constructs a field with the given key and value.
func Int(key string, val int) Field {
	return zap.Int(key, val)
}

// Error is shorthand for the common idiom NamedError("error", err).
func Error(err error) Field {
	return zap.Error(err)
}

// String constructs a field with the given key and value.
func String(key string, val string) Field {
	return zap.String(key, val)
}

// Duration constructs a field with the given key and value
func Duration(key string, val time.Duration) Field {
	return zap.Duration(key, val)
}

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func With(fields ...Field) *Logger {
	return _log.With(fields...)
}
