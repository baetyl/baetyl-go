package crd

// Kind app model kind, crd on k8s
type Kind string

// All kinds
const (
	KindNode          Kind = "node"
	KindApp           Kind = "app"
	KindApplication   Kind = "application"
	KindConfig        Kind = "config"
	KindConfiguration Kind = "configuration"
	KindSecret        Kind = "secret"
)
