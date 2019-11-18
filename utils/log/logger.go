package log

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var log Logger
var level = logrus.InfoLevel

// Logger the logger interface
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

func init() {
	entry := logrus.NewEntry(logrus.New())
	entry.Level = level
	entry.Logger.Out = os.Stdout
	entry.Logger.Level = level
	entry.Logger.Formatter = newFormatter("text")
	log = entry
}

func newFormatter(format string) logrus.Formatter {
	var formatter logrus.Formatter
	if strings.ToLower(format) == "json" {
		formatter = &logrus.JSONFormatter{}
	} else {
		formatter = &logrus.TextFormatter{FullTimestamp: true, DisableColors: true}
	}
	return formatter
}

// Debugf prints log in debug level
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Infof prints log in info level
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Warnf prints log in warn level
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Errorf prints log in error level
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Fatalf prints log in fatal level
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}
