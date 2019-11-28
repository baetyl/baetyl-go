package log

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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
	log.Info("baetyl", Int("age", 12), Error(errors.New("custom error")), String("icon", "baetyl"), Duration("duration", time.Duration(1)))
	log.Sync()
	assert.FileExists(t, jsonFile)

	bytes, err := ioutil.ReadFile(jsonFile)
	assert.NoError(t, err)
	res, _ := regexp.MatchString(`{"level":"info","ts":[0-9T:\.]+,"caller":".*logger_test.*","msg":"baetyl","age":12,"error":"custom error","errorVerbose":".*logger_test.*","icon":"baetyl","duration":.*}`, string(bytes))
	assert.True(t, res)

	log.Error("test error")
	log.Sync()
	bytes, err = ioutil.ReadFile(jsonFile)
	assert.NoError(t, err)
	res, _ = regexp.MatchString(`{"level":"error","ts":[0-9T:\.]+,"caller":".*logger_test.*","msg":"test error","stacktrace":".*"}`, string(bytes))
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
	res, _ = regexp.MatchString(`{"level":"info","ts":[0-9T:\.]+,"caller":".*logger_test.*","msg":"baetyl","name":"baetyl"}`, string(bytes))
	assert.True(t, res)

	cfg.Level = "xxx"
	log = New(cfg)
	log.Debug("baetyl")
	log.Sync()
	bytes, err = ioutil.ReadFile(jsonFile)
	assert.NoError(t, err)
	assert.NotContains(t, string(bytes), `"level":"debug"`)

	log = New(cfg, String("height", "122"))
	assert.NotEmpty(t, log)
	log.Info("baetyl")
	log.Sync()
	bytes, err = ioutil.ReadFile(jsonFile)
	assert.NoError(t, err)
	res, _ = regexp.MatchString(`{"level":"info","ts":[0-9T:\.]+,"caller":".*logger_test.*","msg":"baetyl","height":"122"}`, string(bytes))
	assert.True(t, res)

	textFile := path.Join(dir, "text.log")
	cfg.Format = "text"
	cfg.Path = textFile
	cfg.Level = "info"
	log = New(cfg)

	log.Info("baetyl")
	log.Sync()
	assert.FileExists(t, textFile)

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

	log = New(cfg, String("height", "122"))
	log.Info("baetyl")
	log.Sync()
	bytes, err = ioutil.ReadFile(textFile)
	assert.NoError(t, err)
	assert.Contains(t, string(bytes), `{"height": "122"}`)
}

func TestParseLevel(t *testing.T) {
	level := parseLevel("fatal")
	assert.Equal(t, FatalLevel, level)

	level = parseLevel("panic")
	assert.Equal(t, PanicLevel, level)

	level = parseLevel("error")
	assert.Equal(t, ErrorLevel, level)

	level = parseLevel("warn")
	assert.Equal(t, WarnLevel, level)

	level = parseLevel("warning")
	assert.Equal(t, WarnLevel, level)

	level = parseLevel("info")
	assert.Equal(t, InfoLevel, level)

	level = parseLevel("debug")
	assert.Equal(t, DebugLevel, level)

	level = parseLevel("xxx")
	assert.Equal(t, InfoLevel, level)
}

func TestField(t *testing.T) {
	key := "age"
	m := Int(key, 10)
	assert.Equal(t, key, m.Key)
	assert.Equal(t, int64(10), m.Integer)

	m = Error(errors.New("test"))
	assert.Equal(t, m.Key, "error")

	m = String(key, "baetyl")
	assert.Equal(t, key, m.Key)
	assert.Equal(t, "baetyl", m.String)

	m = Duration(key, time.Duration(12))
	assert.Equal(t, key, m.Key)
	assert.Equal(t, int64(12), m.Integer)
}

func TestNewFileHook(t *testing.T) {
	url := url.URL{
		Scheme: "lumberjack",
		RawQuery: fmt.Sprintf("path=%s&level=%s&format=%s&age_max=%d&size_max=%d&backup_max=%d",
			"test.log", "info", "json", 12, 13, 14),
	}
	lumber, err := newFileHook(&url)
	assert.NoError(t, err)
	assert.True(t, lumber.(*lumberjackSink).Compress)
	assert.Equal(t, "test.log", lumber.(*lumberjackSink).Filename)
	assert.Equal(t, 12, lumber.(*lumberjackSink).MaxAge)
	assert.Equal(t, 13, lumber.(*lumberjackSink).MaxSize)
	assert.Equal(t, 14, lumber.(*lumberjackSink).MaxBackups)
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
