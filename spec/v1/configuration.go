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
}

// ConfigurationObject extended feature for object configuration
type ConfigurationObject struct {
	MD5         string                  `json:"md5,omitempty" yaml:"md5"`
	URL         string                  `json:"url,omitempty" yaml:"url"`
	Compression string                  `json:"compression,omitempty" yaml:"compression"`
	Meta        ConfigurationObjectMeta `json:"meta,omitempty" yaml:"meta"`
}

type ConfigurationObjectMeta struct {
	Kind   string
	Bucket string
	Object string
}
