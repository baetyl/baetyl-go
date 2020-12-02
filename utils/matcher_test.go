package utils

import (
	"testing"

	"gotest.tools/assert"
)

func TestLabelMatcher_Match(t *testing.T) {
	labels := make(map[string]string)
	labels["a"] = "b"
	labels["c"] = "d"

	sl := "a in (b),c=d"
	res, _ := IsLabelMatch(sl, labels)
	assert.Equal(t, true, res)

	sl = "a=bc=d"
	_, err := IsLabelMatch(sl, labels)
	assert.Equal(t, true, err != nil)

	sl = ""
	res, _ = IsLabelMatch(sl, labels)
	assert.Equal(t, true, res)

	var sl1 string
	res, _ = IsLabelMatch(sl1, labels)
	assert.Equal(t, true, res)
}
