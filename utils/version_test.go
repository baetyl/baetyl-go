package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrintVersion(t *testing.T) {
	PrintVersion()
	assert.Equal(t, " Version: unknown\nRevision: unknown", Version())

	VERSION = "1.0.0"
	REVISION = "git-xxx"
	PrintVersion()
	assert.Equal(t, " Version: 1.0.0\nRevision: git-xxx", Version())
}