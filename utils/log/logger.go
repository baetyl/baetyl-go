package log

import (
	"fmt"
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
	FatalLevel
)

var _log, _ = zap.NewDevelopment()

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
func New(c Config, fields ...string) *Logger {
	logLevel, err := parseLevel(c.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse log level (%s), use default level (info)", c.Level)
		logLevel = zapcore.InfoLevel
	}

	fileHook := newFileHook(c)
	encoderConfig := newEncoderConfig()
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(logLevel)
	encoder := newEncoder(c.Format, encoderConfig)
	caller := zap.AddCaller()
	stacktrace := zap.AddStacktrace(zapcore.ErrorLevel)

	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileHook)),
		atomicLevel,
	)

	var fs []Field
	for index := 0; index < len(fields)-1; index = index + 2 {
		f := zap.String(fields[index], fields[index+1])
		fs = append(fs, f)
	}
	_log = zap.New(core, caller, stacktrace, zap.Fields(fs...))
	return _log
}

func parseLevel(lvl string) (zapcore.Level, error) {
	switch strings.ToLower(lvl) {
	case "fatal":
		return zap.FatalLevel, nil
	case "panic":
		return zap.PanicLevel, nil
	case "error":
		return zap.ErrorLevel, nil
	case "warn", "warning":
		return zap.WarnLevel, nil
	case "info":
		return zap.InfoLevel, nil
	case "debug":
		return zap.DebugLevel, nil
	}

	var l zapcore.Level
	return l, fmt.Errorf("not a valid zap level: %q", lvl)
}

func newFileHook(c Config) *lumberjack.Logger {
	var fileHook lumberjack.Logger
	if c.Path != "" {
		err := os.MkdirAll(filepath.Dir(c.Path), 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create log directory: %s", err.Error())
		} else {
			fileHook = lumberjack.Logger{
				Filename:   c.Path,
				MaxSize:    c.Size.Max,
				MaxAge:     c.Age.Max,
				MaxBackups: c.Backup.Max,
				LocalTime:  true,
				Compress:   true,
			}
		}
	}
	return &fileHook
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
		EncodeCaller:   zapcore.FullCallerEncoder,
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
