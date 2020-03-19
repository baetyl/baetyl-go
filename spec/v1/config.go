package v1

import (
	"time"
)

// Configuration
type Configuration struct {
	Name              string            `json:"name,omitempty"`
	Namespace         string            `json:"namespace,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Data              map[string]string `json:"data,omitempty" default:"{}" binding:"required"`
	CreationTimestamp time.Time         `json:"creationTimestamp,omitempty"`
	UpdateTimestamp   time.Time         `json:"updateTimestamp,omitempty"`
	Description       string            `json:"description,omitempty"`
	Version           string            `json:"version,omitempty"`
}
