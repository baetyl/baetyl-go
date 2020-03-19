package http

import (
	gohttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/baetyl/baetyl-go/mock"
	"github.com/baetyl/baetyl-go/utils"
	"github.com/stretchr/testify/assert"
)

func TestClient_Call(t *testing.T) {
	tlssvr, err := utils.NewTLSConfigServer(utils.Certificate{CA: "../mock/testcert/ca.pem", Key: "../mock/testcert/server.key", Cert: "../mock/testcert/server.pem"})
	assert.NoError(t, err)
	assert.NotNil(t, tlssvr)
	tlscli, err := utils.NewTLSConfigClient(utils.Certificate{CA: "../mock/testcert/ca.pem", Key: "../mock/testcert/client.key", Cert: "../mock/testcert/client.pem", InsecureSkipVerify: true})
	assert.NoError(t, err)
	assert.NotNil(t, tlscli)

	ms := mock.NewServer(tlssvr, mock.NewResponse(200, []byte("abc")))
	defer ms.Close()

	ops := NewClientOptions()
	ops.Address = ms.URL
	ops.TLSConfig = tlscli
	c := NewClient(ops)
	resp, err := c.Call("service", "function", []byte("{}"))
	assert.NoError(t, err)
	assert.Equal(t, "abc", string(resp))
}

func TestClieneBadRequest(t *testing.T) {
	ts := httptest.NewServer(gohttp.HandlerFunc(func(w gohttp.ResponseWriter, r *gohttp.Request) {
		w.WriteHeader(gohttp.StatusBadRequest)
	}))
	defer ts.Close()

	ops := NewClientOptions()
	ops.Address = ts.URL
	c := NewClient(ops)

	data, err := c.Call("service", "function", []byte("{}"))
	assert.EqualError(t, err, "[400] 400 Bad Request")
	assert.Empty(t, data)

	resp, err := c.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, gohttp.StatusBadRequest, resp.StatusCode)

	resp, err = c.Post(ts.URL, "", nil)
	assert.NoError(t, err)
	assert.Equal(t, gohttp.StatusBadRequest, resp.StatusCode)
}
