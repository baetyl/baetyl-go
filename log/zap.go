package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/baetyl/baetyl-go/v2/errors"
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

// Code constructs a field with the given value and code key
func Code(err error) Field {
	switch e := err.(type) {
	case errors.Coder:
		return zap.Any("errorCode", e.Code())
	default:
		return zap.Skip()
	}
}

// Error constructs a field with the given value and error key
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
