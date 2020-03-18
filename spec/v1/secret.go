package v1

import "time"

// Secret secret
type Secret struct {
	Name              string            `json:"name,omitempty"`
	Namespace         string            `json:"namespace,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Data              map[string][]byte `json:"data,omitempty"`
	CreationTimestamp time.Time         `json:"creationTimestamp,omitempty"`
	UpdateTimestamp   time.Time         `json:"updateTimestamp,omitempty"`
	Description       string            `json:"description"`
	Version           string            `json:"version,omitempty"`
}