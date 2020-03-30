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

const (
	// baetyl cloud
	SecretLabel    = "secret-type"
	// speical secret of the the registry secret
	SecretRegistry = "baetyl-secret-registry"
	// speical secret of the the config secret
	SecretConfig   = "baetyl-secret-config"
	// speical secret of the the certificate secret
	SecretCertificate   = "baetyl-secret-certificate"
)
