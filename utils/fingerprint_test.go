package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFingerprint(t *testing.T) {
	id1, err := GetFingerprint("")
	assert.NoError(t, err)
	id2, err := GetFingerprint("")
	assert.NoError(t, err)
	assert.Equal(t, id1, id2)
}
