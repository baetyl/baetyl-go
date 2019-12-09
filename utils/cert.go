package utils

import (
	"crypto/tls"
	"net"

	"github.com/256dpi/gomqtt/transport"
	"github.com/docker/go-connections/tlsconfig"
)

// Certificate certificate config for server
// Name : serverNameOverride, same to CommonName in server.pem
// if Name == "" , link would not verifies the server's certificate chain and host name
// AuthType : declares the policy the server will follow for TLS Client Authentication
type Certificate struct {
	CA                 string `yaml:"ca" json:"ca"`
	Key                string `yaml:"key" json:"key"`
	Cert               string `yaml:"cert" json:"cert"`
	Name               string `yaml:"name" json:"name"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify" json:"insecureSkipVerify"` // for client, for test purpose
}

// NewTLSConfigServer loads tls config for server
func NewTLSConfigServer(c *Certificate) (*tls.Config, error) {
	if c == nil || (c.Cert == "" && c.Key == "") {
		return nil, nil
	}
	return tlsconfig.Server(tlsconfig.Options{CAFile: c.CA, KeyFile: c.Key, CertFile: c.Cert, ClientAuth: tls.VerifyClientCertIfGiven})
}

// NewTLSConfigClient loads tls config for client
func NewTLSConfigClient(c *Certificate) (*tls.Config, error) {
	if c == nil {
		return nil, nil
	}
	return tlsconfig.Client(tlsconfig.Options{CAFile: c.CA, KeyFile: c.Key, CertFile: c.Cert, InsecureSkipVerify: c.InsecureSkipVerify})
}

// IsBidirectionalAuthentication check bidirectional authentication
func IsBidirectionalAuthentication(conn transport.Conn) bool {
	var inner net.Conn
	if tcps, ok := conn.(*transport.NetConn); ok {
		inner = tcps.UnderlyingConn()
	} else if wss, ok := conn.(*transport.WebSocketConn); ok {
		inner = wss.UnderlyingConn().UnderlyingConn()
	}
	tlsconn, ok := inner.(*tls.Conn)
	if !ok {
		return false
	}
	state := tlsconn.ConnectionState()
	if !state.HandshakeComplete {
		return false
	}
	return len(state.PeerCertificates) > 0
}
