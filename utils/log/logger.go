package log

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var _log *Logger

func init() {
	var err error
	_log, err = zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to new default logger: %s", err.Error()))
	}
	err = zap.RegisterSink("lumberjack", newFileHook)
	if err != nil {
		_log.Error("failed to register lumberjack", Error(err))
	}
}

// Init init and return logger
func Init(c Config, fields ...Field) (*Logger, error) {
	config := zap.NewProductionConfig()
	if c.Path != "" {
		config.OutputPaths = append(config.OutputPaths, "lumberjack:?"+c.String())
	}
	if c.Format == "text" {
		config.Encoding = "console"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}
	config.Level = zap.NewAtomicLevelAt(parseLevel(c.Level))
	tmp, err := config.Build(zap.Fields(fields...))
	if err != nil {
		return nil, err
	}
	_log = tmp
	return _log, nil
}

type lumberjackSink struct {
	*lumberjack.Logger
}

func (*lumberjackSink) Sync() error {
	return nil
}

func newFileHook(u *url.URL) (zap.Sink, error) {
	args := u.Query()
	path, _ := base64.URLEncoding.DecodeString(args.Get("path"))
	err := os.MkdirAll(filepath.Dir(string(path)), 0755)
	if err != nil {
		_log.Warn("failed to create log directory", Error(err))
		return nil, err
	}
	inner := &lumberjack.Logger{Filename: string(path), Compress: true}
	if age, err := strconv.Atoi(args.Get("age_max")); err == nil {
		inner.MaxAge = age
	}
	if size, err := strconv.Atoi(args.Get("size_max")); err == nil {
		inner.MaxSize = size
	}
	if backup, err := strconv.Atoi(args.Get("backup_max")); err == nil {
		inner.MaxBackups = backup
	}
	return &lumberjackSink{inner}, nil
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
