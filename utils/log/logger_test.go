package log

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	log, err := zap.NewDevelopment()
	assert.NoError(t, err)
	Init(log)

	log = With(String("t", "1"))
	log.Info("failed to xxx", Duration("cost", time.Duration(1)))
	if ent := log.Check(DebugLevel, "xxx"); ent != nil {
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
