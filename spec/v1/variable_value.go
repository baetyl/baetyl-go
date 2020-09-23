package v1

import (
	"encoding/json"
)

// VariableValue variable value which can be app, config or secret
type VariableValue struct {
	Value interface{}
}

// UnmarshalJSON unmarshal from json data
func (v *VariableValue) UnmarshalJSON(b []byte) error {
	v.Value = b
	return nil
}

// MarshalJSON marshal to json data
func (v VariableValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Value)
}

// Unmarshal unmarshal from json data to obj
func (v *VariableValue) Unmarshal(obj interface{}) error {
	switch t := v.Value.(type) {
	case []byte:
		err := json.Unmarshal(t, obj)
		if err != nil {
			return err
		}
		v.Value = obj
	default:
	}
	return nil
}
