package v1

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
	Value        LazyValue `yaml:"value,omitempty" json:"value,omitempty"`
}

// App return app data if its kind is app
func (v *ResourceValue) App() *Application {
	if v.Kind == KindApplication || v.Kind == KindApp {
		var app Application
		v.Value.Unmarshal(&app)
		return &app
	}
	return nil
}

// Config return config data if its kind is config
func (v *ResourceValue) Config() *Configuration {
	if v.Kind == KindConfiguration || v.Kind == KindConfig {
		var cfg Configuration
		v.Value.Unmarshal(&cfg)
		return &cfg
	}
	return nil
}

// Secret return secret data if its kind is secret
func (v *ResourceValue) Secret() *Secret {
	if v.Kind == KindSecret {
		var sec Secret
		v.Value.Unmarshal(&sec)
		return &sec
	}
	return nil
}
