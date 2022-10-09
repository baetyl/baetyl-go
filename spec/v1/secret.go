package v1

import "time"

// Secret secret info
type Secret struct {
	Name              string            `json:"name,omitempty" yaml:"name,omitempty" binding:"resourceName"`
	Namespace         string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Labels            map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Data              map[string][]byte `json:"data,omitempty" yaml:"data,omitempty"`
	CreationTimestamp time.Time         `json:"createTime,omitempty" yaml:"createTime,omitempty"`
	UpdateTimestamp   time.Time         `json:"updateTime,omitempty" yaml:"updateTime,omitempty"`
	Description       string            `json:"description,omitempty" yaml:"description,omitempty"`
	Version           string            `json:"version,omitempty" yaml:"version,omitempty"`
	System            bool              `json:"system,omitempty" yaml:"system,omitempty"`
}
