package crd

import (
	"time"
)

// Secret secret info
type Secret struct {
	Name              string            `json:"name,omitempty"`
	Namespace         string            `json:"namespace,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Data              map[string][]byte `json:"data,omitempty"`
	CreationTimestamp time.Time         `json:"creationTimestamp,omitempty"`
	UpdateTimestamp   time.Time         `json:"updateTimestamp,omitempty"`
	Description       string            `json:"description,omitempty"`
	Version           string            `json:"version,omitempty"`
}

const (
	BaetylCloudGroup = "cloud.baetyl.io"
	SecretLabel      = "secret-type"
	SecretRegistry   = BaetylCloudGroup + "-secret-registry"
	SecretConfig     = BaetylCloudGroup + "-secret-config"
)
