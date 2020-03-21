package api

import (
	"encoding/json"

	"github.com/baetyl/baetyl-go/spec/crd"
)

// CRDRequest body of request to sync crd data
type CRDRequest struct {
	CRDInfos []CRDInfo `yaml:"crds" json:"crds"`
}

// CRDResponse body of response to sync crd data
type CRDResponse struct {
	CRDDatas []CRDData `yaml:"crds" json:"crds"`
}

// CRDInfo crd info
type CRDInfo struct {
	Kind    crd.Kind `yaml:"kind,omitempty" json:"kind,omitempty"`
	Name    string   `yaml:"name,omitempty" json:"name,omitempty"`
	Version string   `yaml:"version,omitempty" json:"version,omitempty"`
}

// CRDData crd data
type CRDData struct {
	CRDInfo `yaml:",inline" json:",inline"`
	Value   VariableValue `yaml:"value,omitempty" json:"value,omitempty"`
}

// App return app crd if kind is app
func (v *CRDData) App() *crd.Application {
	if v.Kind == crd.KindApplication || v.Kind == crd.KindApp {
		if v.Value.Value == nil {
			v.Value.Value = &crd.Application{}
			json.Unmarshal(v.Value.Data, v.Value.Value)
		}
		return v.Value.Value.(*crd.Application)
	}
	return nil
}

// Config return config crd if kind is config
func (v *CRDData) Config() *crd.Configuration {
	if v.Kind == crd.KindConfiguration || v.Kind == crd.KindConfig {
		if v.Value.Value == nil {
			v.Value.Value = &crd.Configuration{}
			json.Unmarshal(v.Value.Data, v.Value.Value)
		}
		return v.Value.Value.(*crd.Configuration)
	}
	return nil
}

// Secret return secret crd if kind is secret
func (v *CRDData) Secret() *crd.Secret {
	if v.Kind == crd.KindSecret {
		if v.Value.Value == nil {
			v.Value.Value = &crd.Secret{}
			json.Unmarshal(v.Value.Data, v.Value.Value)
		}
		return v.Value.Value.(*crd.Secret)
	}
	return nil
}

// VariableValue variable value which can be app, config or secret
type VariableValue struct {
	Data  []byte
	Value interface{}
}

// UnmarshalJSON unmarshal from json data
func (v *VariableValue) UnmarshalJSON(b []byte) error {
	v.Data = b
	return nil
}

// MarshalJSON marshal to json data
func (v *VariableValue) MarshalJSON() ([]byte, error) {
	var err error
	if v.Data == nil && v.Value != nil {
		v.Data, err = json.Marshal(v.Value)
	}
	return v.Data, err
}

// CRDConfigObject extended feature for downloadding object
type CRDConfigObject struct {
	MD5         string `json:"md5,omitempty" yaml:"md5"`
	URL         string `json:"url,omitempty" yaml:"url"`
	Compression string `json:"compression,omitempty" yaml:"compression"`
}
