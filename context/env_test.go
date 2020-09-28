package context

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHostPathLib(t *testing.T) {
	hostPathLib, err := GetHostPathLib()
	assert.NoError(t, err)
	assert.Equal(t, DefaultHostPathLib, hostPathLib)
	os.Setenv(KeyBaetylHostPathLib, "/var/data")
	hostPathLib, err = GetHostPathLib()
	assert.NoError(t, err)
	assert.Equal(t, "/var/data", hostPathLib)
}

