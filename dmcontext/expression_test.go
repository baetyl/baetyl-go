package dmcontext

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseExpression(t *testing.T) {
	expression, err := ParseExpression("")
	assert.Nil(t, expression)
	assert.NoError(t, err)

	expression, err = ParseExpression("sadaaads")
	assert.Nil(t, expression)
	assert.Error(t, err)

	expression, err = ParseExpression("test(x1,x2)")
	assert.Nil(t, expression)
	assert.Error(t, err)

	expression, err = ParseExpression("sum(x1,y2)")
	assert.Nil(t, expression)
	assert.Error(t, err)

	expression, err = ParseExpression("product(x1,x2,10)")
	assert.NotNil(t, expression)
	assert.NoError(t, err)
	assert.Equal(t, MethodProduct, expression.Method)
	assert.Equal(t, 2, len(expression.Args))
	assert.Equal(t, "1", expression.Args[0])
	assert.Equal(t, "2", expression.Args[1])
	assert.Equal(t, 1, len(expression.Nums))
	assert.Equal(t, "10", expression.Nums[0])
}

func TestExecMapping(t *testing.T) {
	res, err := ExecMapping("test", []string{}, "int64")
	assert.Nil(t, res)
	assert.Error(t, err)

	res, err = ExecMapping("equal", []string{}, "int64")
	assert.Nil(t, res)
	assert.Error(t, err)
	res, err = ExecMapping("equal", []string{"1"}, "int64")
	assert.NoError(t, err)
	assert.Equal(t, "1", res)

	res, err = ExecMapping("sum", []string{"1.01"}, "int64")
	assert.Nil(t, res)
	assert.Error(t, err)
	res, err = ExecMapping("sum", []string{"1.01", "x"}, "int64")
	assert.Nil(t, res)
	assert.Error(t, err)
	res, err = ExecMapping("sum", []string{"1.01", "0.99"}, "int64")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), res.(int64))
	res, err = ExecMapping("sum", []string{"1.01", "0.99"}, "float32")
	assert.NoError(t, err)
	assert.Equal(t, float32(2), res.(float32))

	res, err = ExecMapping("product", []string{"1.01"}, "int64")
	assert.Nil(t, res)
	assert.Error(t, err)
	res, err = ExecMapping("product", []string{"1.01", "x"}, "int64")
	assert.Nil(t, res)
	assert.Error(t, err)
	res, err = ExecMapping("product", []string{"1.01", "0.99"}, "int64")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), res.(int64))
	res, err = ExecMapping("product", []string{"1.01", "0.99"}, "float32")
	assert.NoError(t, err)
	assert.Equal(t, float32(0.9999), res.(float32))

	res, err = ExecMapping("subtraction", []string{"1.01"}, "int64")
	assert.Nil(t, res)
	assert.Error(t, err)
	res, err = ExecMapping("subtraction", []string{"1.01", "x"}, "int64")
	assert.Nil(t, res)
	assert.Error(t, err)
	res, err = ExecMapping("subtraction", []string{"1.01", "0.99"}, "int64")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), res.(int64))
	res, err = ExecMapping("subtraction", []string{"1.01", "0.99"}, "float32")
	assert.NoError(t, err)
	assert.Equal(t, float32(0.02), res.(float32))
	res, err = ExecMapping("subtraction", []string{"1.01", "0.99"}, "float64")
	assert.NoError(t, err)
	assert.Equal(t, 0.02, res.(float64))

	res, err = ExecMapping("ratio", []string{"1.01"}, "int64")
	assert.Nil(t, res)
	assert.Error(t, err)
	res, err = ExecMapping("ratio", []string{"1.01", "x"}, "int64")
	assert.Nil(t, res)
	assert.Error(t, err)
	res, err = ExecMapping("ratio", []string{"1.01", "0.99"}, "int64")
	assert.Nil(t, res)
	assert.Error(t, err)
	res, err = ExecMapping("ratio", []string{"1.0201", "1.01"}, "int64")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), res.(int64))
	res, err = ExecMapping("ratio", []string{"1.0201", "1.01"}, "float32")
	assert.NoError(t, err)
	assert.Equal(t, float32(1.01), res.(float32))
	res, err = ExecMapping("ratio", []string{"1.0201", "1.01"}, "float64")
	assert.NoError(t, err)
	assert.Equal(t, 1.01, res.(float64))
}
