package errors

import (
	goerrors "errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	assert.NotEqual(t, New("c", "abc"), New("c", "abc"))
	assert.NotEqual(t, New("c", "abc"), New("c", "xyc"))
	assert.NotEqual(t, New("c", "abc"), New("b", "abc"))

	e := New("c", "abc")
	w := Trace(e)
	assert.Equal(t, e, w)
	assert.EqualError(t, e, "abc")
	assert.EqualError(t, w, "abc")
	assert.Equal(t, e.(Coder).Code(), "c")
	assert.Equal(t, w.(Coder).Code(), "c")
	assert.Contains(t, fmt.Sprintf("%+v", e), "baetyl-go/errors/error.go:30")
	assert.Contains(t, fmt.Sprintf("%+v", w), "baetyl-go/errors/error.go:30")
	_, ok := e.(fmt.Formatter)
	assert.True(t, ok)
	_, ok = w.(fmt.Formatter)
	assert.True(t, ok)

	e2 := goerrors.New("ddd")
	w2 := Trace(e2)
	assert.NotEqual(t, e2, w2)
	assert.EqualError(t, e2, "ddd")
	assert.EqualError(t, w2, "ddd")
	assert.Equal(t, "ddd", fmt.Sprintf("%+v", e2))
	assert.Contains(t, fmt.Sprintf("%+v", w2), "baetyl-go/errors/error.go:20")
	_, ok = e2.(Coder)
	assert.False(t, ok)
	_, ok = w2.(Coder)
	assert.False(t, ok)
	_, ok = e2.(fmt.Formatter)
	assert.False(t, ok)
	_, ok = w2.(fmt.Formatter)
	assert.True(t, ok)

	w3 := Trace(w2)
	assert.Equal(t, w2, w3)

	w4 := Trace(nil)
	assert.Nil(t, w4)
}
