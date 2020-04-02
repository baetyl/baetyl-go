package crd

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
