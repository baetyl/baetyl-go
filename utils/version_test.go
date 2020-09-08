package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintVersion(t *testing.T) {
	PrintVersion()
	assert.Equal(t, " Version: unknown\nRevision: unknown", Version())

	VERSION = "1.0.0"
	REVISION = "git-xxx"
	PrintVersion()
	assert.Equal(t, " Version: 1.0.0\nRevision: git-xxx", Version())
}
