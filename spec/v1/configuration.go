package v1

import (
	"time"
)

// Configuration config info
type Configuration struct {
	Name              string            `json:"name,omitempty" validate:"resourceName,nonBaetyl"`
	Namespace         string            `json:"namespace,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Data              map[string]string `json:"data,omitempty" default:"{}" binding:"required"`
	CreationTimestamp time.Time         `json:"createTime,omitempty"`
	UpdateTimestamp   time.Time         `json:"updateTime,omitempty"`
	Description       string            `json:"description,omitempty"`
	Version           string            `json:"version,omitempty"`
	System            bool              `json:"system,omitempty"`
}

// ConfigurationObject extended feature for object configuration
type ConfigurationObject struct {
	// hex format
	MD5    string `json:"md5,omitempty" yaml:"md5"`
	Sha256 string `json:"sha256,omitempty" yaml:"sha256"`
	URL    string `json:"url,omitempty" yaml:"url"`
	Token  string `json:"token,omitempty" yaml:"token"`
	Unpack string `json:"unpack,omitempty" yaml:"unpack"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata"`
}
