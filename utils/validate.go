package utils

import (
	"reflect"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
)

const (
	resName         = "res_name"
	svcName         = "svc_name"
	fingerprint     = "fingerprint"
	memory          = "memory"
	duration        = "duration"
	devModel        = "dev_model"
	namespace       = "namespace"
	validLabels     = "label"
	validConfigKeys = "config_key"
	dateType        = "data_type"
	dateEnumType    = "enum_type"
	datePlusType    = "data_plus_type"

	nonzero   = "nonzero"
	nonnil    = "nonnil"
	nonbaetyl = "nonbaetyl"
)

var regexps = map[string]string{
	resName:         "^[a-z0-9][-a-z0-9.]{0,61}[a-z0-9]$",
	svcName:         "^[a-z0-9][-a-z0-9]{0,61}[a-z0-9]$",
	fingerprint:     "^[a-zA-Z0-9][-a-zA-Z0-9.]{0,61}[a-zA-Z0-9]$",
	memory:          "^[1-9][0-9]*(k|m|g|t|p|)$",
	duration:        "^[1-9][0-9]*(s|m|h)$",
	devModel:        "^[a-zA-Z0-9\\-_]{1,32}$",
	namespace:       "^[a-z0-9]([-a-z0-9]*[a-z0-9])?([a-z0-9]([-a-z0-9]*[a-z0-9])?)*$",
	validConfigKeys: "^[-._a-zA-Z0-9]+$",
	dateType:        "^(int16|int32|int64|float32|float64|string|bool|time|date)?$",
	dateEnumType:    "^(int16|int32|int64|string)?$",
	datePlusType:    "^(int16|int32|int64|float32|float64|string|time|date|bool|array|enum|object)?$",
}
var validate *validator.Validate

func init() {
	validate = validator.New()
	RegisterValidate(validate)
}

func GetValidator() *validator.Validate {
	return validate
}

func RegisterValidation(key string, fn validator.Func) {
	GetValidator().RegisterValidation(key, fn)
}

func RegisterValidate(v *validator.Validate) {
	if v != nil {
		for key, val := range regexps {
			key0, val0 := key, val
			v.RegisterValidation(key0, func(fl validator.FieldLevel) bool {
				match, _ := regexp.MatchString(val0, fl.Field().String())
				return match
			})
		}

		v.RegisterValidation(nonzero, func(fl validator.FieldLevel) bool {
			return nonzeroValid(fl.Field().Interface())
		})

		v.RegisterValidation(nonnil, func(fl validator.FieldLevel) bool {
			return nonnilValid(fl.Field().Interface())
		})

		v.RegisterValidation(nonbaetyl, func(fl validator.FieldLevel) bool {
			return !strings.Contains(strings.ToLower(fl.Field().String()), "baetyl")
		})

		v.RegisterValidation(validLabels, validLabelsFunc())
	}
}

func validLabelsFunc() validator.Func {
	return func(fl validator.FieldLevel) bool {
		labels, ok := fl.Field().Interface().(map[string]string)
		if !ok {
			return false
		}
		labelRegex, _ := regexp.Compile("^([A-Za-z0-9][-A-Za-z0-9_\\.]*)?[A-Za-z0-9]?$")
		for k, v := range labels {
			if strings.Contains(k, "/") {
				ss := strings.Split(k, "/")
				if len(ss) != 2 {
					return false
				}
				if len(ss[0]) > 253 || len(ss[0]) < 1 || !labelRegex.MatchString(ss[0]) || len(ss[1]) > 63 || !labelRegex.MatchString(ss[1]) {
					return false
				}
			} else {
				if len(k) > 63 || !labelRegex.MatchString(k) {
					return false
				}
			}
			if len(v) > 63 || !labelRegex.MatchString(v) {
				return false
			}
		}
		return true
	}
}

// nonzeroValid tests whether a variable value non-zero as defined by the golang spec.
func nonzeroValid(v interface{}) bool {
	st := reflect.ValueOf(v)
	valid := true
	switch st.Kind() {
	case reflect.String:
		valid = utf8.RuneCountInString(st.String()) != 0
	case reflect.Ptr, reflect.Interface:
		valid = !st.IsNil()
	case reflect.Slice, reflect.Map, reflect.Array:
		valid = st.Len() != 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valid = st.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		valid = st.Uint() != 0
	case reflect.Float32, reflect.Float64:
		valid = st.Float() != 0
	case reflect.Bool:
		valid = st.Bool()
	case reflect.Invalid:
		valid = false // always invalid
	case reflect.Struct:
		valid = true // always valid since only nil pointers are empty
	default:
		valid = false
	}
	return valid
}

// nonnilValid validates that the given pointer is not nil
func nonnilValid(v interface{}) bool {
	st := reflect.ValueOf(v)
	switch st.Kind() {
	case reflect.Ptr, reflect.Interface:
		if st.IsNil() {
			return false
		}
	case reflect.Invalid:
		return false
	}
	return true
}
