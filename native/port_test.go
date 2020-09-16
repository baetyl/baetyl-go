package native

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPortAllocator(t *testing.T) {
	_, err := NewPortAllocator(-1, 2)
	assert.Error(t, err)
	_, err = NewPortAllocator(30000, 1000000)
	assert.Error(t, err)
	_, err = NewPortAllocator(1024, 1024)
	assert.Error(t, err)

	alloc, err := NewPortAllocator(50010, 50011)
	assert.NoError(t, err)
	port1, err := alloc.Allocate()
	assert.NoError(t, err)
	assert.Equal(t, port1, 50010)
	port2, err := alloc.Allocate()
	assert.NoError(t, err)
	assert.Equal(t, port2, 50011)

	listener1, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port1))
	assert.NoError(t, err)

	listener2, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port2))
	assert.NoError(t, err)

	_, err = alloc.Allocate()
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "no available ports in range 50010-50011")

	listener1.Close()
	listener2.Close()

	port4, err := alloc.Allocate()
	assert.NoError(t, err)
	assert.Equal(t, port4, 50010)
}
