package dmcontext

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseValue(t *testing.T) {
	_, err := ParseValue("xx", nil, nil)
	assert.Error(t, err)

	args, err := ParseValue("int16", 1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, args)
	assert.Equal(t, "int16", reflect.TypeOf(args).Name())

	args, err = ParseValue("int32", 1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, args)
	assert.Equal(t, "int32", reflect.TypeOf(args).Name())

	args, err = ParseValue("int64", 1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, args)
	assert.Equal(t, "int64", reflect.TypeOf(args).Name())

	args, err = ParseValue("float32", 1.2, nil)
	assert.NoError(t, err)
	assert.NotNil(t, args)
	assert.Equal(t, "float32", reflect.TypeOf(args).Name())

	args, err = ParseValue("float64", 1.2, nil)
	assert.NoError(t, err)
	assert.NotNil(t, args)
	assert.Equal(t, "float64", reflect.TypeOf(args).Name())

	args, err = ParseValue("string", "1.2", nil)
	assert.NoError(t, err)
	assert.NotNil(t, args)
	assert.Equal(t, "string", reflect.TypeOf(args).Name())

	args, err = ParseValue("time", "17.04.05", "hh:mm:ss")
	assert.NoError(t, err)
	assert.NotNil(t, args)
	assert.Equal(t, "string", reflect.TypeOf(args).Name())

	args, err = ParseValue("time", "17.04.05", "HH:mm:ss")
	assert.NoError(t, err)
	assert.NotNil(t, args)
	assert.Equal(t, "string", reflect.TypeOf(args).Name())

	args, err = ParseValue("time", "17.04.05", "SS:mm:ss")
	assert.Error(t, err)
	assert.Equal(t, "", args)

	ti, _ := time.Parse("2006-01-02", "2022-10-19")
	args, err = ParseValue("time", ti, "yyyy/mm/dd")
	assert.NoError(t, err)
	assert.NotNil(t, args)
	assert.Equal(t, "string", reflect.TypeOf(args).Name())
	assert.Equal(t, "2022/10/19", args)

	arrayType := ArrayType{
		Type: TypeString,
		Min:  2,
		Max:  4,
	}
	testArray := [3]int{1, 2}
	args, err = ParseValue("array", testArray, arrayType)
	assert.NoError(t, err)
	assert.NotNil(t, args)
	arg := args.([]interface{})
	fmt.Println(arg[0])
	assert.Equal(t, "1", arg[0])

	testArray2 := [1]int{1}
	args, err = ParseValue("array", testArray2, arrayType)
	assert.Error(t, err)
	assert.Nil(t, args)

	enum := EnumType{
		Type:   "string",
		Values: []EnumValue{{"1", "Fire", "火灾"}},
	}
	args, err = ParseValue("enum", "Fire", enum)
	assert.NoError(t, err)
	assert.NotNil(t, args)
	assert.Equal(t, "1", args)

	enum1 := EnumType{
		Type:   "xx",
		Values: []EnumValue{{"1", "Fire", "火灾"}},
	}
	args, err = ParseValue("enum", "Fire", enum1)
	assert.Error(t, err)
	assert.Equal(t, "", args)

	ob := map[string]ObjectType{
		"name":  {DisplayName: "demo", Type: "string"},
		"age":   {DisplayName: "demo", Type: "int"},
		"class": {DisplayName: "demo", Type: "string"},
	}
	test := map[string]interface{}{
		"name": "te",
		"age":  12,
	}
	args, err = ParseValue("object", test, ob)
	assert.NoError(t, err)
	assert.NotNil(t, args)
	testMap := args.(map[string]interface{})
	assert.Equal(t, 12, testMap["age"])
}
