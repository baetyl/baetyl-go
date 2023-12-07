package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFingerprint(t *testing.T) {
	id1, _ := GetFingerprint("")
	id2, _ := GetFingerprint("")
	assert.Equal(t, id1, id2)
}
