package dmcontext

import (
	"reflect"
	"strings"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/spf13/cast"
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

func ParseValue(typ string, value, args interface{}) (interface{}, error) {
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

func parseTime(value, args interface{}) (string, error) {
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

func parseArray(value, args interface{}) ([]interface{}, error) {
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
	var newArray []interface{}
	for i := 0; i < originArray.Len(); i++ {
		parseVal, err := ParseValue(arrayType.Type, originArray.Index(i).Interface(), arrayType.Format)
		if err != nil {
			return nil, err
		}
		newArray = append(newArray, parseVal)
	}
	return newArray, nil
}

func parseEnum(value, args interface{}) (string, error) {
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

func parseObject(value, args interface{}) (map[string]interface{}, error) {
	// validate params
	objectTypes, argsOk := args.(map[string]ObjectType)
	originMap, valueOk := value.(map[string]interface{})
	if !argsOk || !valueOk {
		return nil, ErrUnsupportedValueType
	}
	// convert object
	parseMap := map[string]interface{}{}
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
