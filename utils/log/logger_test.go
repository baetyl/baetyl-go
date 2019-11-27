package log

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestLogger(t *testing.T) {
	log := With(String("height", "122"))
	log.Info("test")

	dir, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	jsonFile := path.Join(dir, "json.log")
	cfg := Config{
		Path:   jsonFile,
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

	log = New(cfg)
	log.Info("baetyl")
	log.Sync()
	assert.FileExists(t, jsonFile)

	bytes, err := ioutil.ReadFile(jsonFile)
	assert.NoError(t, err)
	res, _ := regexp.MatchString(`{"level":"info","time":"[0-9T:\.\-\+]+","caller":".*","msg":"baetyl"}`, string(bytes))
	assert.True(t, res)

	log.Error("test error")
	log.Sync()
	bytes, err = ioutil.ReadFile(jsonFile)
	assert.NoError(t, err)
	res, _ = regexp.MatchString(`{"level":"error","time":"[0-9T:\.\-\+]+","caller":".*","msg":"test error","stacktrace":".*"}`, string(bytes))
	assert.True(t, res)

	log.Debug("baetyl")
	log.Sync()
	bytes, err = ioutil.ReadFile(jsonFile)
	assert.NoError(t, err)
	assert.NotContains(t, string(bytes), `"level":"debug"`)

	log = With(String("name", "baetyl"))
	log.Info("baetyl")

	bytes, err = ioutil.ReadFile(jsonFile)
	assert.NoError(t, err)
	res, _ = regexp.MatchString(`{"level":"info","time":"[0-9T:\.\-\+]+","caller":".*","msg":"baetyl","name":"baetyl"}`, string(bytes))
	assert.True(t, res)

	cfg.Level = "xxx"
	log = New(cfg)
	log.Debug("baetyl")
	log.Sync()
	bytes, err = ioutil.ReadFile(jsonFile)
	assert.NoError(t, err)
	assert.NotContains(t, string(bytes), `"level":"debug"`)

	log = New(cfg, "height", "122")
	assert.NotEmpty(t, log)
	log.Info("baetyl")
	log.Sync()
	bytes, err = ioutil.ReadFile(jsonFile)
	assert.NoError(t, err)
	fmt.Println(string(bytes))
	res, _ = regexp.MatchString(`{"level":"info","time":"[0-9T:\.\-\+]+","caller":".+","msg":"baetyl","height":"122"}`, string(bytes))
	assert.True(t, res)

	textFile := path.Join(dir, "text.log")
	cfg.Format = "text"
	cfg.Path = textFile
	log = New(cfg)

	log.Info("baetyl")
	log.Sync()
	assert.FileExists(t, jsonFile)

	bytes, err = ioutil.ReadFile(textFile)
	assert.NoError(t, err)
	assert.Contains(t, string(bytes), "info")

	log.Debug("baetyl")
	log.Sync()
	bytes, err = ioutil.ReadFile(textFile)
	assert.NoError(t, err)
	assert.NotContains(t, string(bytes), "debug")

	log = With(String("name", "baetyl"))
	log.Info("baetyl")
	bytes, err = ioutil.ReadFile(textFile)
	assert.NoError(t, err)
	assert.Contains(t, string(bytes), `{"name": "baetyl"}`)

	cfg.Level = "xxx"
	log = New(cfg)
	log.Debug("baetyl")
	log.Sync()
	bytes, err = ioutil.ReadFile(textFile)
	assert.NoError(t, err)
	assert.NotContains(t, string(bytes), "debug")

	log = New(cfg, "height", "122")
	log.Info("baetyl")
	log.Sync()
	bytes, err = ioutil.ReadFile(textFile)
	assert.NoError(t, err)
	assert.Contains(t, string(bytes), `{"height": "122"}`)
}

func TestParseLevel(t *testing.T) {
	level, err := parseLevel("fatal")
	assert.NoError(t, err)
	assert.Equal(t, zapcore.FatalLevel, level)

	level, err = parseLevel("panic")
	assert.NoError(t, err)
	assert.Equal(t, zapcore.PanicLevel, level)

	level, err = parseLevel("error")
	assert.NoError(t, err)
	assert.Equal(t, zapcore.ErrorLevel, level)

	level, err = parseLevel("warn")
	assert.NoError(t, err)
	assert.Equal(t, zapcore.WarnLevel, level)

	level, err = parseLevel("warning")
	assert.NoError(t, err)
	assert.Equal(t, zapcore.WarnLevel, level)

	level, err = parseLevel("info")
	assert.NoError(t, err)
	assert.Equal(t, zapcore.InfoLevel, level)

	level, err = parseLevel("debug")
	assert.NoError(t, err)
	assert.Equal(t, zapcore.DebugLevel, level)

	level, err = parseLevel("xxx")
	assert.Error(t, err)
}

func TestField(t *testing.T) {
	key := "age"
	m := Int(key, 10)
	assert.Equal(t, key, m.Key)
	assert.Equal(t, int64(10), m.Integer)

	m = Error(errors.New("test"))
	assert.Equal(t, m.Key, "error")
	assert.Equal(t, zapcore.ErrorType, m.Type)

	m = String(key, "baetyl")
	assert.Equal(t, key, m.Key)
	assert.Equal(t, "baetyl", m.String)

	m = Duration(key, time.Duration(12))
	assert.Equal(t, key, m.Key)
	assert.Equal(t, zapcore.DurationType, m.Type)
	assert.Equal(t, int64(12), m.Integer)
}

func BenchmarkConsoleAndFile(b *testing.B) {
	dir, err := ioutil.TempDir("", b.Name())
	assert.NoError(b, err)
	defer os.RemoveAll(dir)

	file := path.Join(dir, "test.log")
	cfg := Config{
		Path:   file,
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
			Max: 1000,
		},
		Backup: struct {
			Max int `yaml:"max" json:"max" default:"15" validate:"min=1"`
		}{
			Max: 15,
		},
	}
	logger := New(cfg)
	b.ResetTimer()
	b.Run("log", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Info("test: " + strconv.Itoa(i))
		}
		logger.Sync()
	})
}
