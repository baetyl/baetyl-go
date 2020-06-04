package errors

import (
	goerrors "errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	assert.NotEqual(t, CodeError("c", "abc"), CodeError("c", "abc"))
	assert.NotEqual(t, CodeError("c", "abc"), CodeError("c", "xyc"))
	assert.NotEqual(t, CodeError("c", "abc"), CodeError("b", "abc"))

	e := CodeError("c", "abc")
	assert.Equal(t, e, Trace(e))
	assert.EqualError(t, e, "abc")
	assert.Equal(t, e.(Coder).Code(), "c")
	assert.Contains(t, fmt.Sprintf("%+v", e), "baetyl-go/errors/error.go:34")
	_, ok := e.(fmt.Formatter)
	assert.True(t, ok)

	e2 := goerrors.New("ddd")
	w2 := Trace(e2)
	assert.NotEqual(t, e2, w2)
	assert.EqualError(t, e2, "ddd")
	assert.EqualError(t, w2, "ddd")
	assert.Equal(t, "ddd", fmt.Sprintf("%+v", e2))
	assert.Contains(t, fmt.Sprintf("%+v", w2), "baetyl-go/errors/error.go:29")
	_, ok = e2.(Coder)
	assert.False(t, ok)
	_, ok = w2.(Coder)
	assert.False(t, ok)
	_, ok = e2.(fmt.Formatter)
	assert.False(t, ok)
	_, ok = w2.(fmt.Formatter)
	assert.True(t, ok)

	e3 := New("edc")
	e4 := Errorf("rfv%s", "1")
	assert.Equal(t, e3, Trace(e3))
	assert.Equal(t, e4, Trace(e4))
	assert.Contains(t, fmt.Sprintf("%+v", e3), "baetyl-go/errors/error.go:14")
	assert.Contains(t, fmt.Sprintf("%+v", e4), "baetyl-go/errors/error.go:18")

	assert.Nil(t, Trace(nil))
}
