package context

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectPlatform(t *testing.T) {
	assert.Equal(t, Platform().OS, runtime.GOOS)
	assert.Contains(t, PlatformString(), runtime.GOOS)
}
