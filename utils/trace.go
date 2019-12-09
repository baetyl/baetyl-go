package utils

import (
	"time"

	"github.com/baetyl/baetyl-go/log"
)

// Trace print elapsed time
func Trace(f func(string, ...log.Field), msg string, fields ...log.Field) func() {
	start := time.Now()
	return func() {
		fields := append(fields, log.Duration("cost", time.Since(start)))
		f(msg, fields...)
	}
}
