package utils

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTLSConfigServer(t *testing.T) {
	tl, err := NewTLSConfigServer(Certificate{Key: "../example/var/lib/baetyl/testcert/server.key", ClientAuthType: tls.VerifyClientCertIfGiven})
	assert.Error(t, err)

	tl, err = NewTLSConfigServer(Certificate{Cert: "../example/var/lib/baetyl/testcert/server.crt"})
	assert.Error(t, err)

	c := Certificate{
		Key:  "../example/var/lib/baetyl/testcert/server.key",
		Cert: "../example/var/lib/baetyl/testcert/server.crt",
	}

	tl, err = NewTLSConfigServer(c)
	assert.NoError(t, err)
	assert.NotEmpty(t, tl)
}

func TestNewTLSConfigClient(t *testing.T) {
	tl, err := NewTLSConfigClient(Certificate{Key: "../example/var/lib/baetyl/testcert/client.key"})
	assert.Error(t, err)

	tl, err = NewTLSConfigClient(Certificate{Cert: "../example/var/lib/baetyl/testcert/client.crt"})
	assert.Error(t, err)
	assert.Empty(t, tl)

	c := Certificate{
		Key:  "../example/var/lib/baetyl/testcert/client.key",
		Cert: "../example/var/lib/baetyl/testcert/client.crt",
	}
	tl, err = NewTLSConfigClient(c)
	assert.NoError(t, err)
	assert.NotEmpty(t, tl)
}
