package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/docker/go-units"
	"gopkg.in/yaml.v2"

	"github.com/baetyl/baetyl-go/v2/errors"
)

// LoadYAML config into out interface, with defaults and validates
func LoadYAML(path string, out interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Trace(err)
	}
	res, err := ParseEnv(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config parse error: %s", err.Error())
		res = data
	}
	return UnmarshalYAML(res, out)
}

// ParseEnv parses env
func ParseEnv(data []byte) ([]byte, error) {
	text := string(data)
	envs := os.Environ()
	envMap := make(map[string]string)
	for _, s := range envs {
		t := strings.Split(s, "=")
		envMap[t[0]] = t[1]
	}
	tmpl, err := template.New("template").Option("missingkey=error").Parse(text)
	if err != nil {
		return nil, errors.Trace(err)
	}
	buffer := bytes.NewBuffer(nil)
	err = tmpl.Execute(buffer, envMap)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return buffer.Bytes(), nil
}

// UnmarshalYAML unmarshals, defaults and validates
func UnmarshalYAML(in []byte, out interface{}) error {
	err := yaml.Unmarshal(in, out)
	if err != nil {
		return errors.Trace(err)
	}
	err = SetDefaults(out)
	if err != nil {
		return errors.Trace(err)
	}
	err = GetValidator().Struct(out)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// UnmarshalJSON unmarshals, defaults and validates
func UnmarshalJSON(in []byte, out interface{}) error {
	err := json.Unmarshal(in, out)
	if err != nil {
		return errors.Trace(err)
	}
	err = SetDefaults(out)
	if err != nil {
		return errors.Trace(err)
	}
	err = GetValidator().Struct(out)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// Size int64
type Size int64

// MarshalYAML customizes marshal
func (s Size) MarshalYAML() (interface{}, error) {
	return int64(s), nil
}

// UnmarshalYAML customizes unmarshal
func (s *Size) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return errors.Trace(err)
	}
	v, err := units.RAMInBytes(str)
	if err != nil {
		return errors.Trace(err)
	}
	*s = Size(v)
	return nil
}

// MarshalJSON customizes marshal
func (s Size) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(s), 10)), nil
}

// UnmarshalJSON customizes unmarshal
func (s *Size) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" {
		return nil
	}
	str = strings.Trim(str, "\"")
	v, err := units.RAMInBytes(str)
	if err != nil {
		return errors.Trace(err)
	}
	*s = Size(v)
	return nil
}

/*
  "b" represents for "B"
  "k" represents for "KB" or "KiB"
  "m" represents for "MB" or "MiB"
  "g" represents for "GB" or "GiB"
  "t" represents for "TB" or "TiB"
  "p" represents for "PB" or "PiB"
  maxValue is (2 >> 63 -1).
*/
var decimapAbbrs = []string{"", "k", "m", "g", "t", "p"}

// Length int64
// ! Length is deprecated, please to use Size
type Length struct {
	Max int64 `yaml:"max" json:"max"`
}

// UnmarshalYAML customizes unmarshal
func (l *Length) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ls length
	err := unmarshal(&ls)
	if err != nil {
		return errors.Trace(err)
	}
	if ls.Max != "" {
		l.Max, err = units.RAMInBytes(ls.Max)
		if err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

// MarshalYAML implements the Marshaller interface
func (l *Length) MarshalYAML() (interface{}, error) {
	var ls length
	ls.Max = units.CustomSize("%.4g%s", float64(l.Max), 1024.0, decimapAbbrs)
	return ls, nil
}

type length struct {
	Max string `yaml:"max" json:"max"`
}
