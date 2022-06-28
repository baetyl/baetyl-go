package dmcontext

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseExpression(t *testing.T) {
	args, err := ParseExpression("")
	assert.NoError(t, err)
	assert.Nil(t, args)

	args, err = ParseExpression("&***)")
	assert.Error(t, err)
	assert.Nil(t, args)

	args, err = ParseExpression("x1--x2")
	assert.Error(t, err)
	assert.Nil(t, args)

	args, err = ParseExpression("4/(1+2+1*3*10)")
	assert.NoError(t, err)
	assert.NotNil(t, args)
	assert.Equal(t, 0, len(args))

	args, err = ParseExpression("x1")
	assert.NoError(t, err)
	assert.NotNil(t, args)
	assert.Equal(t, 1, len(args))
	assert.Equal(t, "x1", args[0])

	args, err = ParseExpression("x4/(x1+x2+x1*x3*10)")
	assert.NoError(t, err)
	assert.NotNil(t, args)
	assert.Equal(t, 5, len(args))
	assert.Equal(t, "x1", args[1])
}

func TestExecExpression(t *testing.T) {
	res, err := ExecExpression("", map[string]interface{}{}, "test")
	assert.Error(t, err)
	assert.Nil(t, res)

	res, err = ExecExpression("", map[string]interface{}{}, MappingNone)
	assert.NoError(t, err)
	assert.Nil(t, res)

	res, err = ExecExpression("", map[string]interface{}{}, MappingValue)
	assert.Error(t, err)
	assert.Nil(t, res)
	res, err = ExecExpression("x1++x2", map[string]interface{}{}, MappingValue)
	assert.Error(t, err)
	assert.Nil(t, res)
	res, err = ExecExpression("x1+x2", map[string]interface{}{}, MappingValue)
	assert.Error(t, err)
	assert.Nil(t, res)
	res, err = ExecExpression("x1", map[string]interface{}{"x2": 1}, MappingValue)
	assert.Error(t, err)
	assert.Nil(t, res)
	res, err = ExecExpression("x1", map[string]interface{}{"x1": 1}, MappingValue)
	assert.NoError(t, err)
	assert.Equal(t, 1, res)
	res, err = ExecExpression("x1", map[string]interface{}{"x1": "1"}, MappingValue)
	assert.NoError(t, err)
	assert.Equal(t, "1", res)

	res, err = ExecExpression("", map[string]interface{}{}, MappingCalculate)
	assert.Error(t, err)
	assert.Nil(t, res)
	res, err = ExecExpression("x1++x2", map[string]interface{}{}, MappingCalculate)
	assert.Error(t, err)
	assert.Nil(t, res)
	res, err = ExecExpression("x4/(x1+x2+x1*x3*10)", map[string]interface{}{"x1": 1}, MappingCalculate)
	assert.Error(t, err)
	assert.Nil(t, res)
	res, err = ExecExpression("x4/(x1+x2+x1*x3*10)", map[string]interface{}{"x1": "asasd"}, MappingCalculate)
	assert.Error(t, err)
	assert.Nil(t, res)
	args := map[string]interface{}{"x1": int16(1), "x2": float32(1.1), "x3": 0.99, "x4": int64(15)}
	res, err = ExecExpression("x4/(x1+x2+x1*x3*10)", args, MappingCalculate)
	assert.NoError(t, err)
	assert.Equal(t, 1.25, res)
}

func TestSolveExpression(t *testing.T) {
	zero := float64(0)

	res, err := SolveExpression("", 1)
	assert.Error(t, err)
	assert.Equal(t, zero, res)

	res, err = SolveExpression("x1*2+x2*3", 1)
	assert.Error(t, err)
	assert.Equal(t, zero, res)

	res, err = SolveExpression("(x1+2)*x1", 1)
	assert.Error(t, err)
	assert.Equal(t, zero, res)

	res, err = SolveExpression("1/(x1+2)", 1)
	assert.Error(t, err)
	assert.Equal(t, zero, res)

	res, err = SolveExpression("(x1+2)&x1", 1)
	assert.Error(t, err)
	assert.Equal(t, zero, res)

	res, err = SolveExpression("x1*2-x1*2+1", 1)
	assert.Error(t, err)
	assert.Equal(t, zero, res)

	res, err = SolveExpression("x1*2-11", 9)
	assert.NoError(t, err)
	assert.Equal(t, float64(10), res)

	res, err = SolveExpression("(x1+1)*3+x1*2+1", 9)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), res)
}
