package trigger

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func A(a, b int) struct{ A int } {
	return struct{ A int }{A: a + b}
}

func B(a, b int) {
	fmt.Println(a + b)
}

func Test_eventTrigger_Exec(t1 *testing.T) {
	err := Register("addFunc", EventFunc{
		Args:  []interface{}{1},
		Event: A,
	})
	assert.NoError(t1, err)

	c, err := Exec("addFunc", 2)
	assert.NoError(t1, err)
	assert.Equal(t1, struct {
		A int
	}{A: 3}, c[0].Interface().(struct{ A int }))

	_, err = Exec("addFunc")
	assert.NotNil(t1, err)

	_, err = Exec("deleteFuc", 1, 2)
	assert.NotNil(t1, err)

	err = Register("addFunc2", EventFunc{
		Args:  []interface{}{1},
		Event: 1,
	})
	assert.NotNil(t1, err)

	err = Register("addFuncB", EventFunc{
		Args:  []interface{}{},
		Event: B,
	})
	assert.NoError(t1, err)
	rs, err := Exec("addFuncB", 1, 2)
	assert.NoError(t1, err)
	assert.Nil(t1, rs)

	SyncExec("addFuncB", 2, 4)
}
