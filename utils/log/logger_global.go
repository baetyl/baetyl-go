package log

import (
	"time"

	"go.uber.org/zap"
)

// Global global logger
var Global *Logger

func init() {
	Global, _ = zap.NewProduction()
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
	return Global.With(fields...)
}
