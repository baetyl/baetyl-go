package context

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectRunMode(t *testing.T) {
	os.Setenv(KeyRunMode, "native")
	assert.Equal(t, "native", RunMode())
	os.Setenv(KeyRunMode, "xxx")
	assert.Equal(t, "kube", RunMode())
}
