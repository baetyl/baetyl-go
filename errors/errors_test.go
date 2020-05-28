package errors

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrors(t *testing.T) {
	assert.NotEqual(t, New("c", "abc"), New("c", "abc"))
	assert.NotEqual(t, New("c", "abc"), New("c", "xyc"))
	assert.NotEqual(t, New("c", "abc"), New("b", "abc"))
	e := New("c", "abc")
	w := Wrap(e, "c", "rfv")
	assert.Equal(t, e, e)
	assert.NotEqual(t, e, w)
	assert.EqualError(t, e, "abc")
	assert.EqualError(t, w, "rfv: abc")
	assert.Equal(t, e.(*CodeError).Code(), "c")
	assert.Equal(t, w.(*CodeError).Code(), "c")
	assert.EqualError(t, e.(*CodeError).Unwrap(), "abc")
	assert.EqualError(t, w.(*CodeError).Unwrap(), "rfv: abc")
}
