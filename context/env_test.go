package context

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHostPathLib(t *testing.T) {
	hostPathLib, err := HostPathLib()
	assert.NoError(t, err)
	assert.Equal(t, defaultHostPathLib, hostPathLib)
	os.Setenv(KeyBaetylHostPathLib, "/var/data")
	hostPathLib, err = HostPathLib()
	assert.NoError(t, err)
	assert.Equal(t, "/var/data", hostPathLib)
}

func TestDetectRunMode(t *testing.T) {
	os.Setenv(KeyRunMode, "native")
	assert.Equal(t, "native", RunMode())
	os.Setenv(KeyRunMode, "xxx")
	assert.Equal(t, "kube", RunMode())
}
