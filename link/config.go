package link

import "time"

// LServerCert certificate config for gRPC server
type LServerCert struct {
	Key  string `yaml:"key" json:"key"`
	Cert string `yaml:"cert" json:"cert"`
}

// LClientCert certificate config for gRPC client
// Name : serverNameOverride, same to CommonName in server.pem
type LClientCert struct {
	Name     string `yaml:"name" json:"name"`
	Cert     string `yaml:"cert" json:"cert"`
	Insecure bool   `yaml:"insecure" json:"insecure"` // for client, for svr purpose
}

// Account authentication information
type Account struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

// LServerConfig link server config
type LServerConfig struct {
	Address     string      `yaml:"address" json:"address" default: "0.0.0.0"`
	Account     Account     `yaml:"account" json:"account"`
	Certificate LServerCert `yaml:",inline" json:",inline"`
}

// LClientConfig link client config
type LClientConfig struct {
	Address     string        `yaml:"address" json:"address" default: "0.0.0.0"`
	Timeout     time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	Account     Account       `yaml:"account" json:"account"`
	Certificate LClientCert   `yaml:",inline" json:",inline"`
}