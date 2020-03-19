package http

import (
	gohttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/baetyl/baetyl-go/mock"
	"github.com/baetyl/baetyl-go/utils"
	"github.com/stretchr/testify/assert"
)

func TestClientRequests(t *testing.T) {
	tlssvr, err := utils.NewTLSConfigServer(utils.Certificate{CA: "../mock/testcert/ca.pem", Key: "../mock/testcert/server.key", Cert: "../mock/testcert/server.pem"})
	assert.NoError(t, err)
	assert.NotNil(t, tlssvr)

	response := mock.NewResponse(200, []byte("abc"))
	ms := mock.NewServer(tlssvr, response, response, response)
	defer ms.Close()

	var cfg ClientConfig
	utils.UnmarshalYAML(nil, &cfg)
	cfg.CA = "../mock/testcert/ca.pem"
	cfg.Key = "../mock/testcert/client.key"
	cfg.Cert = "../mock/testcert/client.pem"
	cfg.InsecureSkipVerify = true
	cfg.Address = ms.URL
	ops, err := cfg.ToClientOptions()
	assert.NoError(t, err)
	assert.NotNil(t, ops)
	c := NewClient(ops)
	resp, err := c.Call("service", "function", []byte("{}"))
	assert.NoError(t, err)
	assert.Equal(t, "abc", string(resp))

	data, err := c.PostJSON("v1", []byte("{}"))
	assert.NoError(t, err)
	assert.Equal(t, "abc", string(data))

	data, err = c.GetJSON("v1")
	assert.NoError(t, err)
	assert.Equal(t, "abc", string(data))
}

func TestClieneBadRequests(t *testing.T) {
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
