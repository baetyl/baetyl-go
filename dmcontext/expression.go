package dmcontext

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"strconv"

	"github.com/crsmithdev/goexpr"

	"github.com/baetyl/baetyl-go/v2/errors"
)

var (
	ErrUnknownMappingType   = errors.New("unknown mapping type")
	ErrUnsupportedValueType = errors.New("unsupported value type")
)

// ParseExpression parse expression string to args
// for example, input: x4/(x1+x2+x1*x3*10), output: [x4,x1,x2,x1,x3]
func ParseExpression(e string) ([]string, error) {
	if e == "" {
		return nil, nil
	}
	expression, err := goexpr.Parse(e)
	if err != nil {
		return nil, errors.Trace(err)
	}
	vars := make([]string, 0)
	if expression.Vars != nil {
		vars = expression.Vars
	}
	return vars, nil
}

// ExecExpression execute expression with args and mappingType
// for example, input: ("x1+x2", '{"x1":1,"x2":2}', "calc"), output: 3
func ExecExpression(e string, args map[string]interface{}, mappingType string) (interface{}, error) {
	switch mappingType {
	case MappingNone:
		return nil, nil
	case MappingValue:
		return processValueMapping(e, args)
	case MappingCalculate:
		return processCalcMapping(e, args)
	default:
		return nil, ErrUnknownMappingType
	}
}

func processValueMapping(e string, args map[string]interface{}) (interface{}, error) {
	// parse expression
	expression, err := goexpr.Parse(e)
	if err != nil {
		return nil, errors.Trace(err)
	}
	// check the number of variables
	if len(expression.Vars) != 1 {
		return nil, errors.New("mapping type equal can only have one variable")
	}
	// check variable exist
	if val, ok := args[expression.Vars[0]]; ok {
		return val, nil
	}
	return nil, errors.New("missing argument:" + expression.Vars[0])
}

func processCalcMapping(e string, args map[string]interface{}) (interface{}, error) {
	// parse expression
	expression, err := goexpr.Parse(e)
	if err != nil {
		return nil, errors.Trace(err)
	}
	// parse variable to float64
	parseArgs := map[string]float64{}
	for _, v := range expression.Vars {
		if _, ok := args[v]; !ok {
			return nil, errors.New("missing variable:" + v)
		}
		val, err := parseValueToFloat64(args[v])
		if err != nil {
			return nil, err
		}
		parseArgs[v] = val
	}
	// calculate result
	res, err := goexpr.Evaluate(expression, parseArgs)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return strconv.ParseFloat(fmt.Sprintf("%.4f", res), 64)
}

func parseValueToFloat64(v interface{}) (float64, error) {
	switch v.(type) {
	case int:
		return float64(v.(int)), nil
	case int16:
		return float64(v.(int16)), nil
	case int32:
		return float64(v.(int32)), nil
	case int64:
		return float64(v.(int64)), nil
	case float32:
		return float64(v.(float32)), nil
	case float64:
		return v.(float64), nil
	default:
		return 0, ErrUnsupportedValueType
	}
}

// SolveExpression solve the expression with value
// Note: currently only support the expression that can be simplified to ax+b
// for example, input: ((x1+1)*3+x1*2+1, 9) which means (x1+1)*3+x1*2+1=9, output: 1 which means x1=1
func SolveExpression(e string, value float64) (float64, error) {
	// parse expression
	expression, err := goexpr.Parse(e)
	if err != nil {
		return 0, errors.Trace(err)
	}
	// check the number of variables
	set := map[string]interface{}{}
	for _, v := range expression.Vars {
		set[v] = nil
	}
	if len(set) != 1 {
		return 0, errors.New("the number of variables in expression is not one")
	}
	// simple expression
	slope, offset, err := simpleExpression(expression.Ast)
	if err != nil {
		return 0, errors.Trace(err)
	}
	// solve expression
	if slope == 0 {
		return 0, errors.New("the slope is zero after simple")
	}
	return (value - offset) / slope, nil
}

// simpleExpression simple node to slope and offset
func simpleExpression(node ast.Node) (float64, float64, error) {
	switch node.(type) {
	case *ast.Ident:
		return 1, 0, nil
	case *ast.BinaryExpr:
		return processBinaryExpr(node)
	case *ast.ParenExpr:
		return simpleExpression(node.(*ast.ParenExpr).X)
	case *ast.BasicLit:
		offset, err := strconv.ParseFloat(node.(*ast.BasicLit).Value, 64)
		if err != nil {
			return 0, 0, err
		}
		return 0, offset, nil
	default:
		return 0, 0, errors.Errorf("unsupported node %+v (type %+v)", node, reflect.TypeOf(node))
	}
}

func processBinaryExpr(node ast.Node) (float64, float64, error) {
	n := node.(*ast.BinaryExpr)
	xa, xb, err := simpleExpression(n.X)
	if err != nil {
		return 0, 0, err
	}
	ya, yb, err := simpleExpression(n.Y)
	if err != nil {
		return 0, 0, err
	}
	switch n.Op {
	case token.ADD:
		return xa + ya, xb + yb, nil
	case token.SUB:
		return xa - ya, xb - yb, nil
	case token.MUL:
		if xa != 0 && ya != 0 {
			return 0, 0, errors.New("only support linear equation")
		}
		return xa*yb + xb*ya, xb * yb, nil
	case token.QUO:
		if ya != 0 {
			return 0, 0, errors.New("denominator can not have a variable")
		}
		return xa / yb, xb / yb, nil
	default:
		return 0, 0, errors.Errorf("unsupported binary operation: %s", n.Op)
	}
}
