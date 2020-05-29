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

// TODO：MD5 using []byte
// ConfigurationObject extended feature for object configuration
type ConfigurationObject struct {
	MD5         string                      `json:"md5,omitempty" yaml:"md5"`
	URL         string                      `json:"url,omitempty" yaml:"url"`
	Compression string                      `json:"compression,omitempty" yaml:"compression"`
	Metadata    ConfigurationObjectMetadata `json:"metadata,omitempty" yaml:"metadata"`
}

type ConfigurationObjectMetadata struct {
	Source string `json:"source,omitempty" yaml:"source"`
	Bucket string `json:"bucket,omitempty" yaml:"bucket"`
	Object string `json:"object,omitempty" yaml:"object"`
	Token  string `json:"token,omitempty" yaml:"token"`
}
