package v1

import "time"

// Secret secret info
type Secret struct {
	Name              string            `json:"name,omitempty" validate:"resourceName,nonBaetyl"`
	Namespace         string            `json:"namespace,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Data              map[string][]byte `json:"data,omitempty"`
	CreationTimestamp time.Time         `json:"createTime,omitempty"`
	UpdateTimestamp   time.Time         `json:"updateTime,omitempty"`
	Description       string            `json:"description,omitempty"`
	Version           string            `json:"version,omitempty"`
	System            bool              `json:"system,omitempty"`
}
