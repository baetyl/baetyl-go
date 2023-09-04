package dmcontext

import (
	"reflect"
	"strings"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/spf13/cast"
)

const (
	TypeInt     = "int"
	TypeInt16   = "int16"
	TypeInt32   = "int32"
	TypeInt64   = "int64"
	TypeFloat32 = "float32"
	TypeFloat64 = "float64"
	TypeBool    = "bool"
	TypeString  = "string"
	TypeTime    = "time"
	TypeDate    = "date"
	TypeArray   = "array"
	TypeEnum    = "enum"
	TypeObject  = "object"
)

var (
	ErrUnsupportedValueType = errors.New("unsupported value type")
)

var timeFormats = map[string]string{
	"yyyy-mm-dd": "2006-01-02",
	"yyyy.mm.dd": "2006.01.02",
	"yyyy/mm/dd": "2006/01/02",
	"mm-dd-yyyy": "06-01-2020",
	"hh:mm:ss":   "15:04:05",
	"HH:MM:SS":   "15:04:05",
}

var parseLayout = [6]string{
	"2006-01-02", "2006.01.02", "2006/01/02", "15:04:05", "15-04-05", "15.04.05",
}

func ParsePropertyValue(tpy string, val float64) (any, error) {
	switch tpy {
	case TypeInt16:
		return int16(val), nil
	case TypeInt32:
		return int32(val), nil
	case TypeInt64:
		return int64(val), nil
	case TypeFloat32:
		return float32(val), nil
	case TypeFloat64:
		return val, nil
	default:
		return nil, ErrTypeNotSupported
	}
}

func ParseValue(typ string, value, args any) (any, error) {
	switch typ {
	case TypeInt:
		return cast.ToIntE(value)
	case TypeInt16:
		return cast.ToInt16E(value)
	case TypeInt32:
		return cast.ToInt32E(value)
	case TypeInt64:
		return cast.ToInt64E(value)
	case TypeFloat32:
		return cast.ToFloat32E(value)
	case TypeFloat64:
		return cast.ToFloat64E(value)
	case TypeBool:
		return cast.ToBoolE(value)
	case TypeString:
		return cast.ToStringE(value)
	case TypeDate, TypeTime:
		return parseTime(value, args)
	case TypeArray:
		return parseArray(value, args)
	case TypeEnum:
		return parseEnum(value, args)
	case TypeObject:
		return parseObject(value, args)
	default:
		return nil, errors.New("unsupported type: " + typ)
	}
}

func parseTime(value, args any) (string, error) {
	// validate params
	format, argsOk := args.(string)
	timeFormat, formatOk := timeFormats[strings.ToLower(format)]
	if !argsOk || !formatOk {
		return "", ErrUnsupportedValueType
	}
	// format time
	switch reflect.TypeOf(value).Name() {
	case TypeString:
		val := value.(string)
		for _, layout := range parseLayout {
			t, err := time.ParseInLocation(layout, val, time.Local)
			if err != nil {
				continue
			}
			return t.Format(timeFormat), nil
		}
		return "", ErrUnsupportedValueType
	case "Time":
		return value.(time.Time).Format(timeFormat), nil
	default:
		return "", ErrUnsupportedValueType
	}
}

func parseArray(value, args any) ([]any, error) {
	// validate params
	arrayType, ok := args.(ArrayType)
	if ok && reflect.TypeOf(value).Kind() != reflect.Array {
		return nil, ErrUnsupportedValueType
	}
	originArray := reflect.ValueOf(value)
	if originArray.Len() > arrayType.Max || originArray.Len() < arrayType.Min {
		return nil, errors.New("the length of the array does not conform to the range")
	}
	// convert array
	var newArray []any
	for i := 0; i < originArray.Len(); i++ {
		parseVal, err := ParseValue(arrayType.Type, originArray.Index(i).Interface(), arrayType.Format)
		if err != nil {
			return nil, err
		}
		newArray = append(newArray, parseVal)
	}
	return newArray, nil
}

func parseEnum(value, args any) (string, error) {
	// validate params
	enumType, ok := args.(EnumType)
	if ok && reflect.TypeOf(value).Name() != enumType.Type {
		return "", ErrUnsupportedValueType
	}
	// convert to enum
	for _, v := range enumType.Values {
		enumValue, err := ParseValue(enumType.Type, v.Value, nil)
		if err != nil {
			return "", err
		}
		if enumValue == value {
			return v.Name, nil
		}
	}
	return "", errors.New("no matching enum value")
}

func parseObject(value, args any) (map[string]any, error) {
	// validate params
	objectTypes, argsOk := args.(map[string]ObjectType)
	originMap, valueOk := value.(map[string]any)
	if !argsOk || !valueOk {
		return nil, ErrUnsupportedValueType
	}
	// convert object
	parseMap := map[string]any{}
	for key, objectType := range objectTypes {
		originVal, ok := originMap[key]
		if !ok {
			continue
		}
		parseValue, err := ParseValue(objectType.Type, originVal, objectType.Format)
		if err != nil {
			return nil, err
		}
		parseMap[key] = parseValue
	}
	return parseMap, nil
}

func ParseValueToFloat64(v any) (float64, error) {
	switch i := v.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint:
		return float64(i), nil
	default:
		return 0, ErrUnsupportedValueType
	}
}

func ParseValueToBool(v interface{}) (bool, error) {
	switch v.(type) {
	case bool:
		return v.(bool), nil
	default:
		return false, ErrUnsupportedValueType
	}
}

func ParseValueToUint32(v interface{}) (uint32, error) {
	switch v.(type) {
	case int16:
		return uint32(v.(int16)), nil
	case int32:
		return uint32(v.(int32)), nil
	case int64:
		return uint32(v.(int64)), nil
	case float32:
		return uint32(v.(float32)), nil
	case float64:
		return uint32(v.(float64)), nil
	default:
		return 0, ErrUnsupportedValueType
	}
}

func ParseValueToFloat32(v interface{}) (float32, error) {
	switch v.(type) {
	case int16:
		return float32(v.(int16)), nil
	case int32:
		return float32(v.(int32)), nil
	case int64:
		return float32(v.(int64)), nil
	case float32:
		return v.(float32), nil
	case float64:
		return float32(v.(float64)), nil
	default:
		return 0, ErrUnsupportedValueType
	}
}
