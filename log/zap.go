package log

import (
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
	PanicLevel
	FatalLevel
)

// Any constructs a field with the given key and value
func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}

// Error is shorthand for the common idiom NamedError("error", err).
func Error(err error) Field {
	return zap.Error(err)
}

// L returns the global Logger, which can be reconfigured with ReplaceGlobals.
// It's safe for concurrent use.
func L() *Logger {
	return zap.L()
}

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func With(fields ...Field) *Logger {
	return zap.L().With(fields...)
}
