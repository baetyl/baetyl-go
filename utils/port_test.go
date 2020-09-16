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

	listener1, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", got))
	assert.NoError(t, err)

	res := CheckPortAvailable("127.0.0.1", got)
	assert.False(t, res)

	listener1.Close()

	res = CheckPortAvailable("127.0.0.1", got)
	assert.True(t, res)
}
