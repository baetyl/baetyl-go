package log

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestLogger(t *testing.T) {
	cfg := Config{
		Path:   "test.log",
		Level:  "info",
		Format: "json",
		Age: struct {
			Max int `yaml:"max" json:"max" default:"15" validate:"min=1"`
		}{
			Max: 15,
		},
		Size: struct {
			Max int `yaml:"max" json:"max" default:"50" validate:"min=1"`
		}{
			Max: 1,
		},
		Backup: struct {
			Max int `yaml:"max" json:"max" default:"15" validate:"min=1"`
		}{
			Max: 15,
		},
	}

	log := InitLogger(cfg)
	assert.NotEmpty(t, log)

	log = With(String("t", "1"))
	log.Info("failed to xxx", Duration("cost", time.Duration(1)))
	if ent := log.Check(zapcore.DebugLevel, "xxx"); ent != nil {
		ent.Write(Int("c", 1))
	}
	log.Error("failed to do", Error(do2()))
}

func do1() error {
	return errors.New("1")
}

func do2() error {
	return do1()
}
