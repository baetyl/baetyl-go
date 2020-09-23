package v1

import (
	"encoding/json"
)

// VariableValue variable value which can be app, config or secret
type LazyValue struct {
	Value interface{}
}

// UnmarshalJSON unmarshal from json data
func (v *LazyValue) UnmarshalJSON(b []byte) error {
	v.Value = b
	return nil
}

// MarshalJSON marshal to json data
func (v LazyValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Value)
}

// Unmarshal unmarshal from json data to obj
func (v *LazyValue) Unmarshal(obj interface{}) error {
	switch t := v.Value.(type) {
	case []byte:
		err := json.Unmarshal(t, obj)
		if err != nil {
			return err
		}
	default:
		data, err := json.Marshal(v.Value)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, obj)
	}
	return nil
}
