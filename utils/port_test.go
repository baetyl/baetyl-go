package utils

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPortAvailable(t *testing.T) {
	got, err := GetAvailablePort("127.0.0.1")
	assert.NoError(t, err)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", got))
	assert.NoError(t, err)
	listener.Close()

	got, err = GetAvailablePort("0.0.0.0")
	assert.NoError(t, err)
	listener, err = net.Listen("tcp", fmt.Sprintf(":%d", got))
	assert.NoError(t, err)
	listener.Close()
}

func TestCheckPortAvailable(t *testing.T) {
	got, err := GetAvailablePort("127.0.0.1")
	assert.NoError(t, err)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", got))
	assert.NoError(t, err)

	res := CheckPortAvailable(got)
	assert.False(t, res)

	listener.Close()

	res = CheckPortAvailable(got)
	assert.True(t, res)
}
