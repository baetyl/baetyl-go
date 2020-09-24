package v1

import (
	"encoding/json"
)

// VariableValue variable value which can be app, config or secret
type LazyValue struct {
	Value interface{}
	doc   []byte
}

// UnmarshalJSON unmarshal from json data
func (v *LazyValue) UnmarshalJSON(b []byte) error {
	v.doc = b
	return nil
}

// SetJSON set the json doc
func (v *LazyValue) SetJSON(doc []byte) {
	v.doc = doc
}

// MarshalJSON marshal to json data
func (v LazyValue) MarshalJSON() ([]byte, error) {
	if v.doc != nil {
		return v.doc, nil
	}
	return json.Marshal(v.Value)
}

// Unmarshal unmarshal from json data to obj
func (v *LazyValue) Unmarshal(obj interface{}) error {
	if v.doc != nil {
		return json.Unmarshal(v.doc, obj)
	}
	if v.Value != nil {
		bs, err := json.Marshal(v.Value)
		if err != nil {
			return err
		}
		return json.Unmarshal(bs, obj)
	}
	return nil
}
