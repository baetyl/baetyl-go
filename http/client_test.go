package http

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl-go/v2/mock"
	"github.com/baetyl/baetyl-go/v2/utils"
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
	resp, err := c.Call("service/function", []byte("{}"))
	assert.NoError(t, err)
	assert.Equal(t, "abc", string(resp))

	data, err := c.PostJSON("v1", []byte("{}"))
	assert.NoError(t, err)
	assert.Equal(t, "abc", string(data))

	data, err = c.GetJSON("v1")
	assert.NoError(t, err)
	assert.Equal(t, "abc", string(data))
}

func TestClientBadRequests(t *testing.T) {
	response := mock.NewResponse(400, []byte("abc"))
	ms := mock.NewServer(nil, response, response, response)
	defer ms.Close()

	ops := NewClientOptions()
	ops.Address = ms.URL
	c := NewClient(ops)

	data, err := c.Call("service/function", []byte("{}"))
	assert.EqualError(t, err, "[400] abc")
	assert.Equal(t, "abc", string(data))

	data, err = c.GetJSON(ms.URL)
	assert.EqualError(t, err, "[400] abc")
	assert.Equal(t, "abc", string(data))

	data, err = c.PostJSON(ms.URL, []byte("abc"))
	assert.EqualError(t, err, "[400] abc")
	assert.Equal(t, "abc", string(data))

	ops.SyncMaxConcurrency = 10
	c = NewClient(ops)
	result := make(chan *SyncResults, 1000)
	for i := 0; i < 100; i++ {
		c.SyncSendUrl("GET", ms.URL, nil, result, map[string]interface{}{})
	}
	time.Sleep(time.Second * 2)
	assert.Equal(t, len(result), 100)

}

func TestSendURL(t *testing.T) {
	resp := []*mock.Response{
		mock.NewResponse(200, []byte("Get")),
		mock.NewResponse(200, []byte("Post")),
		mock.NewResponse(200, []byte("Put")),
		mock.NewResponse(200, []byte("Put")),
		mock.NewResponse(200, []byte("Delete")),
	}
	ms := mock.NewServer(nil, resp...)
	defer ms.Close()

	header := map[string]string{
		"a": "b",
	}

	cli := NewClient(NewClientOptions())
	res, err := cli.GetURL(ms.URL, header)
	assert.NoError(t, err)
	data, err := HandleResponse(res)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Get"), data)

	res, err = cli.PostURL(ms.URL, bytes.NewReader([]byte("body")), header)
	assert.NoError(t, err)
	data, err = HandleResponse(res)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Post"), data)

	res, err = cli.GetURL(ms.URL, header)
	assert.NoError(t, err)
	data, err = HandleResponse(res)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Put"), data)

	res, err = cli.PutURL(ms.URL, bytes.NewReader([]byte("body")), header)
	assert.NoError(t, err)
	data, err = HandleResponse(res)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Put"), data)

	res, err = cli.DeleteURL(ms.URL, header)
	assert.NoError(t, err)
	data, err = HandleResponse(res)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Delete"), data)
}
