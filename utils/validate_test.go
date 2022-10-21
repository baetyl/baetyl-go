package utils

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type testValidate struct {
	Res string      `json:"res" validate:"res_name"`
	Svc string      `json:"svc" validate:"svc_name"`
	Fp  string      `json:"fp" validate:"fingerprint"`
	Mem string      `json:"mem" validate:"memory"`
	Dur string      `json:"dur" validate:"duration"`
	Dm  string      `json:"dm" validate:"dev_model"`
	Nb  string      `json:"nb" validate:"nonbaetyl"`
	Ns  string      `json:"ns" validate:"namespace"`
	Ll  interface{} `json:"ll" validate:"label"`
	Cfg string      `json:"cfg" validate:"config_key"`
	Dt  string      `json:"dt" validate:"data_type"`
	Et  string      `json:"et" validate:"enum_type"`
	Dpt string      `json:"dpt" validate:"data_plus_type"`
	Le  string      `json:"le" validate:"min=1,max=5"`
	Nn  string      `json:"nn" validate:"nonzero"`
	Ptr *int        `json:"ptr" validate:"nonzero"`
	It  interface{} `json:"it" validate:"nonnil"`
	Om  string      `json:"om" validate:"omitempty"`
	Req string      `json:"req" validate:"required"`
}

type customValidate struct {
	Custom string `json:"custom" validate:"custom"`
}

func TestGetValidator(t *testing.T) {
	v := GetValidator()
	assert.NotNil(t, v)
}

func TestRegisterValidation(t *testing.T) {
	RegisterValidation("custom", func(fl validator.FieldLevel) bool {
		s := fl.Field().Interface().(string)
		return len(s) < 5
	})

	c := customValidate{Custom: "test"}
	err := GetValidator().Struct(c)
	assert.NoError(t, err)

	c.Custom = "test1"
	err = GetValidator().Struct(c)
	assert.Error(t, err)
}

func TestRegisterValidate(t *testing.T) {
	v := validator.New()
	RegisterValidate(v)

	x := 1

	c := testValidate{
		Res: "test",
		Svc: "qwe",
		Fp:  "qwe",
		Mem: "123g",
		Dur: "123s",
		Dm:  "qwe",
		Nb:  "qwe",
		Ns:  "qwe",
		Ll:  map[string]string{"a": "b"},
		Cfg: "qwe",
		Dt:  "int64",
		Et:  "string",
		Dpt: "date",
		Le:  "123",
		Nn:  "qwe",
		Ptr: &x,
		It:  3,
		Req: "qwe",
	}

	err := v.Struct(c)
	assert.NoError(t, err)

	c.Res = "-ads"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Res = ""
	err = v.Struct(c)
	assert.Error(t, err)

	c.Res = "QWE"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Res = "0123456789012345678901234567890123456789012345678901234567890123" // len=64
	err = v.Struct(c)
	assert.Error(t, err)

	c.Res = "ads.123"
	err = v.Struct(c)
	assert.NoError(t, err)

	c.Svc = "-ads"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Svc = ""
	err = v.Struct(c)
	assert.Error(t, err)

	c.Svc = "QWE"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Svc = "0123456789012345678901234567890123456789012345678901234567890123" // len=64
	err = v.Struct(c)
	assert.Error(t, err)

	c.Svc = "ads"
	err = v.Struct(c)
	assert.NoError(t, err)

	c.Fp = "-ads"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Fp = "0123456789012345678901234567890123456789012345678901234567890123" // len=64
	err = v.Struct(c)
	assert.Error(t, err)

	c.Fp = ""
	err = v.Struct(c)
	assert.Error(t, err)

	c.Fp = "adsDDD.ewew213FF"
	err = v.Struct(c)
	assert.NoError(t, err)

	c.Mem = "123cc"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Mem = "123g"
	err = v.Struct(c)
	assert.NoError(t, err)

	c.Dur = "123s321"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Dur = "123123h"
	err = v.Struct(c)
	assert.NoError(t, err)

	c.Dm = "-Soiewq123@"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Dm = "S2wce"
	err = v.Struct(c)
	assert.NoError(t, err)

	c.Nb = "qwebaetyl000"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Nb = "qwebAeTyl000"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Nb = "adsDDD"
	err = v.Struct(c)
	assert.NoError(t, err)

	c.Ns = "nas.123.EE"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Ns = "default"
	err = v.Struct(c)
	assert.NoError(t, err)

	c.Ll = "wqeeqw.EEwq#"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Ll = nil
	err = v.Struct(c)
	assert.Error(t, err)

	c.Ll = map[string]string{"a": "b"}
	err = v.Struct(c)
	assert.NoError(t, err)

	c.Cfg = "#12"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Cfg = "ad-eqw.asd"
	err = v.Struct(c)
	assert.NoError(t, err)

	c.Dt = "err"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Dt = "int32"
	err = v.Struct(c)
	assert.NoError(t, err)

	c.Et = "hehe"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Et = "string"
	err = v.Struct(c)
	assert.NoError(t, err)

	c.Dpt = "datetime22"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Dpt = "time"
	err = v.Struct(c)
	assert.NoError(t, err)

	c.Le = ""
	err = v.Struct(c)
	assert.Error(t, err)

	c.Le = "123123"
	err = v.Struct(c)
	assert.Error(t, err)

	c.Le = "time"
	err = v.Struct(c)
	assert.NoError(t, err)

	c.Nn = ""
	err = v.Struct(c)
	assert.Error(t, err)

	c.Nn = "nn"
	err = v.Struct(c)
	assert.NoError(t, err)

	c.Ptr = nil
	err = v.Struct(c)
	assert.Error(t, err)

	c.Ptr = &x
	err = v.Struct(c)
	assert.NoError(t, err)

	c.It = nil
	err = v.Struct(c)
	assert.Error(t, err)

	c.It = "it"
	err = v.Struct(c)
	assert.NoError(t, err)

	c.Req = ""
	err = v.Struct(c)
	assert.Error(t, err)

	c.Req = "req"
	err = v.Struct(c)
	assert.NoError(t, err)
}
