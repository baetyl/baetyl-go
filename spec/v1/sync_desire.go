package v1

import (
	"encoding/json"
)

// DesireRequest body of request to sync desired data
type DesireRequest struct {
	Infos []ResourceInfo `yaml:"infos" json:"infos"`
}

// DesireResponse body of response to sync desired data
type DesireResponse struct {
	Values []ResourceValue `yaml:"values" json:"values"`
}

// ResourceInfo desired info
type ResourceInfo struct {
	Kind    Kind   `yaml:"kind,omitempty" json:"kind,omitempty"`
	Name    string `yaml:"name,omitempty" json:"name,omitempty"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
}

// ResourceValue desired value
type ResourceValue struct {
	ResourceInfo `yaml:",inline" json:",inline"`
	Value        VariableValue `yaml:"value,omitempty" json:"value,omitempty"`
}

// App return app data if its kind is app
func (v *ResourceValue) App() *Application {
	if v.Kind == KindApplication || v.Kind == KindApp {
		if v.Value.Value == nil {
			v.Value.Value = &Application{}
			json.Unmarshal(v.Value.data, v.Value.Value)
		}
		return v.Value.Value.(*Application)
	}
	return nil
}

// Config return config data if its kind is config
func (v *ResourceValue) Config() *Configuration {
	if v.Kind == KindConfiguration || v.Kind == KindConfig {
		if v.Value.Value == nil {
			v.Value.Value = &Configuration{}
			json.Unmarshal(v.Value.data, v.Value.Value)
		}
		return v.Value.Value.(*Configuration)
	}
	return nil
}

// Secret return secret data if its kind is secret
func (v *ResourceValue) Secret() *Secret {
	if v.Kind == KindSecret {
		if v.Value.Value == nil {
			v.Value.Value = &Secret{}
			json.Unmarshal(v.Value.data, v.Value.Value)
		}
		return v.Value.Value.(*Secret)
	}
	return nil
}
