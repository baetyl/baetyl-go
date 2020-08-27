package v1

import (
	"encoding/json"
)

// VariableValue variable value which can be app, config or secret
type VariableValue struct {
	data  []byte
	Value interface{}
}

// UnmarshalJSON unmarshal from json data
func (v *VariableValue) UnmarshalJSON(b []byte) error {
	v.data = b
	return nil
}

// MarshalJSON marshal to json data
func (v VariableValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Value)
}
