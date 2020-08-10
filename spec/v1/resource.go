package v1

// Kind app model kind, crd on k8s
type Kind string

// MessageKind message kind
type MessageKind string

// All kinds
const (
	KindNode          Kind = "node"
	KindApp           Kind = "app"
	KindApplication   Kind = "application"
	KindConfig        Kind = "config"
	KindConfiguration Kind = "configuration"
	KindSecret        Kind = "secret"
)

const (
	SecretLabel = "secret-type"
	// speical secret of the the registry
	SecretRegistry = "registry"
	// speical secret of the the config
	SecretConfig = "config"
	// speical secret of the the certificate
	SecretCertificate = "certificate"
)

const (
	MessageReport MessageKind = "report"
	MessageDesire MessageKind = "desire"
)
