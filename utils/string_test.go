package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandString(t *testing.T) {
	res := RandString(6)
	assert.Equal(t, len(res), 6)
}
