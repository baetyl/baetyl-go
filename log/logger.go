package log

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/baetyl/baetyl-go/v2/errors"
)

func init() {
	// Config{
	// 	Level:       NewAtomicLevelAt(InfoLevel),
	// 	Development: false,
	// 	Sampling: &SamplingConfig{
	// 		Initial:    100,
	// 		Thereafter: 100,
	// 	},
	// 	Encoding:         "json",
	// 	EncoderConfig:    NewProductionEncoderConfig(),
	// 	OutputPaths:      []string{"stderr"},
	// 	ErrorOutputPaths: []string{"stderr"},
	// }
	c := zap.NewProductionConfig()
	c.Sampling = nil
	c.OutputPaths = []string{"stdout"}
	l, err := c.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to create default logger: %s", err.Error()))
	}
	err = zap.RegisterSink("lumberjack", newFileHook)
	if err != nil {
		l.Error("failed to register lumberjack", Error(err))
	}
	zap.ReplaceGlobals(l)
}

// Init init and return logger
func Init(cfg Config, fields ...Field) (*Logger, error) {
	c := zap.NewProductionConfig()
	c.Sampling = nil
	if cfg.Filename != "" {
		c.OutputPaths = append(c.OutputPaths, "lumberjack:?"+cfg.String())
	}
	if cfg.Encoding == "console" {
		c.Encoding = "console"
		c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}
	if cfg.EncodeTime != "" {
		c.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(cfg.EncodeTime))
		}
	}
	if cfg.EncodeLevel != "" {
		c.EncoderConfig.EncodeLevel = func(lvl zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			ft := strings.ReplaceAll(cfg.EncodeLevel, "level", "%s")
			enc.AppendString(fmt.Sprintf(ft, lvl.String()))
		}
	}
	c.Level = zap.NewAtomicLevelAt(parseLevel(cfg.Level))
	l, err := c.Build(zap.Fields(fields...))
	if err != nil {
		return nil, errors.Trace(err)
	}
	zap.ReplaceGlobals(l)
	return L(), nil
}

type lumberjackSink struct {
	*lumberjack.Logger
}

func (*lumberjackSink) Sync() error {
	return nil
}

func newFileHook(u *url.URL) (zap.Sink, error) {
	cfg, err := FromURL(u)
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = os.MkdirAll(filepath.Dir(cfg.Filename), 0755)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &lumberjackSink{&lumberjack.Logger{
		Compress:   cfg.Compress,
		Filename:   cfg.Filename,
		MaxAge:     cfg.MaxAge,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
	}}, nil
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
		L().Warn("failed to parse log level, use default level (info)", Any("level", lvl))
		return InfoLevel
	}
}
