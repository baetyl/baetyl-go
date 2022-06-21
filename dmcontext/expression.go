package dmcontext

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/baetyl/baetyl-go/v2/errors"
)

const (
	MethodEqual       = "equal"
	MethodSum         = "sum"
	MethodProduct     = "product"
	MethodSubtraction = "subtraction"
	MethodRatio       = "ratio"
)

var (
	ErrInvalidExpression       = errors.New("invalid expression")
	ErrInvalidExpressionArgs   = errors.New("invalid expression args")
	ErrUnknownExpressionMethod = errors.New("unknown expression method")
	ErrUnsupportedArgType      = errors.New("unsupported arg type")
	ErrDivisorZero             = errors.New("divisor can not be zero")
)

type Expression struct {
	Method string   `yaml:"method,omitempty" json:"method,omitempty"`
	Args   []string `yaml:"args,omitempty" json:"args,omitempty"`
	Nums   []string `yaml:"nums,omitempty" json:"nums,omitempty"`
}

// ParseExpression parse expression string to expression struct
func ParseExpression(e string) (*Expression, error) {
	if e == "" {
		return nil, nil
	}

	index := strings.Index(e, "(")
	if index < 0 {
		return nil, ErrInvalidExpression
	}

	method := e[:index]
	switch method {
	case MethodEqual, MethodSum, MethodProduct, MethodSubtraction, MethodRatio:
		var args, nums []string
		originArgs := strings.Split(e[index+1:len(e)-1], ",")
		for _, arg := range originArgs {
			if strings.HasPrefix(arg, "x") {
				args = append(args, arg[1:])
			} else if _, err := strconv.ParseFloat(arg, 64); err == nil {
				nums = append(nums, arg)
			} else {
				return nil, ErrInvalidExpressionArgs
			}
		}
		return &Expression{
			Method: method,
			Args:   args,
			Nums:   nums,
		}, nil
	default:
		return nil, ErrUnknownExpressionMethod
	}
}

// ExecMapping mapping result by method and args
// when method is equal, resType not work, return the first arg directly
func ExecMapping(method string, args []string, resType string) (interface{}, error) {
	// mapping result according to method
	switch method {
	case MethodEqual:
		if len(args) != 1 {
			return nil, errors.New("method equal, the number of args is not one")
		}
		return args[0], nil
	case MethodSum:
		if len(args) < 2 {
			return nil, errors.New("method sum, the number of args less than two")
		}
		// parse args to float64 array
		parseArgs, err := parseToFloat64(args)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return processSum(parseArgs, resType)
	case MethodProduct:
		if len(args) < 2 {
			return nil, errors.New("method product, the number of args less than two")
		}
		// parse args to float64 array
		parseArgs, err := parseToFloat64(args)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return processProduct(parseArgs, resType)
	case MethodSubtraction:
		if len(args) != 2 {
			return nil, errors.New("method subtraction, the number of args is not two")
		}
		// parse args to float64 array
		parseArgs, err := parseToFloat64(args)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return processSubtraction(parseArgs, resType)
	case MethodRatio:
		if len(args) != 2 {
			return nil, errors.New("method ratio, the number of args is not two")
		}
		// parse args to float64 array
		parseArgs, err := parseToFloat64(args)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return processRatio(parseArgs, resType)
	default:
		return nil, ErrUnknownExpressionMethod
	}
}

func parseToFloat64(args []string) ([]float64, error) {
	var parse []float64
	for _, arg := range args {
		a, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			return nil, err
		}
		parse = append(parse, a)
	}
	return parse, nil
}

func processSum(args []float64, resType string) (interface{}, error) {
	switch resType {
	case TypeInt16:
		res := int16(0)
		for _, arg := range args {
			res += int16(arg)
		}
		return res, nil
	case TypeInt32:
		res := int32(0)
		for _, arg := range args {
			res += int32(arg)
		}
		return res, nil
	case TypeInt64:
		res := int64(0)
		for _, arg := range args {
			res += int64(arg)
		}
		return res, nil
	case TypeFloat32:
		res := float32(0)
		for _, arg := range args {
			res += float32(arg)
		}
		return res, nil
	case TypeFloat64:
		res := float64(0)
		for _, arg := range args {
			res += arg
		}
		return res, nil
	default:
		return nil, ErrUnsupportedArgType
	}
}

func processProduct(args []float64, resType string) (interface{}, error) {
	switch resType {
	case TypeInt16:
		res := int16(1)
		for _, arg := range args {
			res *= int16(arg)
		}
		return res, nil
	case TypeInt32:
		res := int32(1)
		for _, arg := range args {
			res *= int32(arg)
		}
		return res, nil
	case TypeInt64:
		res := int64(1)
		for _, arg := range args {
			res *= int64(arg)
		}
		return res, nil
	case TypeFloat32:
		res := float32(1)
		for _, arg := range args {
			res *= float32(arg)
		}
		return res, nil
	case TypeFloat64:
		res := float64(1)
		for _, arg := range args {
			res *= arg
		}
		return res, nil
	default:
		return nil, ErrUnsupportedArgType
	}
}

func processSubtraction(args []float64, resType string) (interface{}, error) {
	switch resType {
	case TypeInt16:
		return int16(args[0]) - int16(args[1]), nil
	case TypeInt32:
		return int32(args[0]) - int32(args[1]), nil
	case TypeInt64:
		return int64(args[0]) - int64(args[1]), nil
	case TypeFloat32:
		res, err := strconv.ParseFloat(fmt.Sprintf("%.4f", args[0]-args[1]), 32)
		if err != nil {
			return nil, err
		}
		return float32(res), nil
	case TypeFloat64:
		return strconv.ParseFloat(fmt.Sprintf("%.4f", args[0]-args[1]), 64)
	default:
		return nil, ErrUnsupportedArgType
	}
}

func processRatio(args []float64, resType string) (interface{}, error) {
	switch resType {
	case TypeInt16:
		dividend, divisor := int16(args[0]), int16(args[1])
		if divisor == 0 {
			return nil, ErrDivisorZero
		}
		return dividend / divisor, nil
	case TypeInt32:
		dividend, divisor := int32(args[0]), int32(args[1])
		if divisor == 0 {
			return nil, ErrDivisorZero
		}
		return dividend / divisor, nil
	case TypeInt64:
		dividend, divisor := int64(args[0]), int64(args[1])
		if divisor == 0 {
			return nil, ErrDivisorZero
		}
		return dividend / divisor, nil
	case TypeFloat32:
		dividend, divisor := float32(args[0]), float32(args[1])
		if divisor == 0 {
			return nil, ErrDivisorZero
		}
		return dividend / divisor, nil
	case TypeFloat64:
		dividend, divisor := args[0], args[1]
		if divisor == 0 {
			return nil, ErrDivisorZero
		}
		return dividend / divisor, nil
	default:
		return nil, ErrUnsupportedArgType
	}
}
