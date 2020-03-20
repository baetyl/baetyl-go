package v1

// Kind app model kind, crd on k8s
type Kind string

// All kinds
const (
	KindNode          Kind = "node"
	KindApp           Kind = "app"
	KindApplication   Kind = "application"
	KindCfg           Kind = "cfg"
	KindConfig        Kind = "config"
	KindConfiguration Kind = "configuration"
	KindSecret        Kind = "secret"
)
