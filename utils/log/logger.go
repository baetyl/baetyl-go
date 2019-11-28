package log

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
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

var _log, _ = zap.NewProduction()

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

// New new logger
func New(c Config, fields ...Field) *Logger {
	logLevel := parseLevel(c.Level)
	fileHook := newFileHook(c)
	encoderConfig := newEncoderConfig()
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(logLevel)
	encoder := newEncoder(c.Format, encoderConfig)
	caller := zap.AddCaller()
	stacktrace := zap.AddStacktrace(WarnLevel)

	var writer zapcore.WriteSyncer
	if fileHook != nil {
		writer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileHook))
	} else {
		writer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))
	}
	core := zapcore.NewCore(
		encoder,
		writer,
		atomicLevel,
	)

	_log = zap.New(core, caller, stacktrace, zap.Fields(fields...))
	return _log
}

func parseLevel(lvl string) Level {
	switch strings.ToLower(lvl) {
	case "fatal":
		return FatalLevel
	case "panic":
		return PanicLevel
	case "error":
		return ErrorLevel
	case "warn", "warning":
		return WarnLevel
	case "info":
		return InfoLevel
	case "debug":
		return DebugLevel
	default:
		_log.Warn("failed to parse log level, use default level (info)", String("level", lvl))
		return InfoLevel
	}
}

func newFileHook(c Config) *lumberjack.Logger {
	if c.Path == "" {
		return nil
	}

	err := os.MkdirAll(filepath.Dir(c.Path), 0755)
	if err != nil {
		_log.Warn("failed to create log directory", Error(err))
		return nil
	}
	return &lumberjack.Logger{
		Filename:   c.Path,
		MaxSize:    c.Size.Max,
		MaxAge:     c.Age.Max,
		MaxBackups: c.Backup.Max,
		Compress:   true,
	}
}

func newEncoderConfig() zapcore.EncoderConfig {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
	return encoderConfig
}

func newEncoder(format string, config zapcore.EncoderConfig) zapcore.Encoder {
	var encoder zapcore.Encoder
	if strings.ToLower(format) == "json" {
		encoder = zapcore.NewJSONEncoder(config)
	} else {
		encoder = zapcore.NewConsoleEncoder(config)
	}
	return encoder
}
