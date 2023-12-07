package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFingerprint(t *testing.T) {
	id1, err1 := GetFingerprint("")
	id2, err2 := GetFingerprint("")
	assert.Equal(t, err1, err2)
	assert.Equal(t, id1, id2)
}
