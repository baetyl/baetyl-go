package v1

import (
	"strings"
	"time"
)

const PrefixConfigObject = "_object_"

// Configuration config info
type Configuration struct {
	Name              string            `json:"name,omitempty" yaml:"name,omitempty" binding:"res_name"`
	Namespace         string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Labels            map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Data              map[string]string `json:"data,omitempty" yaml:"data,omitempty" default:"{}" binding:"required"`
	CreationTimestamp time.Time         `json:"createTime,omitempty" yaml:"createTime,omitempty"`
	UpdateTimestamp   time.Time         `json:"updateTime,omitempty" yaml:"updateTime,omitempty"`
	Description       string            `json:"description,omitempty" yaml:"description,omitempty"`
	Version           string            `json:"version,omitempty" yaml:"version,omitempty"`
	System            bool              `json:"system,omitempty" yaml:"system,omitempty"`
}

// ConfigurationObject extended feature for object configuration
type ConfigurationObject struct {
	// hex format
	MD5      string            `json:"md5,omitempty" yaml:"md5,omitempty"`
	Sha256   string            `json:"sha256,omitempty" yaml:"sha256,omitempty"`
	URL      string            `json:"url,omitempty" yaml:"url,omitempty"`
	Token    string            `json:"token,omitempty" yaml:"token,omitempty"`
	Unpack   string            `json:"unpack,omitempty" yaml:"unpack,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

func IsConfigObject(key string) bool {
	return strings.HasPrefix(key, PrefixConfigObject)
}
