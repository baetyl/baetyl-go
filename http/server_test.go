package http

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/baetyl/baetyl-go/v2/utils"
)

func TestServerHttp(t *testing.T) {
	confs := []struct {
		serverConf ServerConfig
		cliConf    mockClientConf
	}{
		{
			serverConf: ServerConfig{
				Address:     "127.0.0.1:50050",
				ReadTimeout: time.Second,
			},
			cliConf: mockClientConf{
				Address: "http://127.0.0.1:50050",
			},
		},
		{
			serverConf: ServerConfig{
				Address:     "127.0.0.1:50060",
				ReadTimeout: time.Second,
				Certificate: utils.Certificate{
					Cert: "../example/var/lib/baetyl/testcert/server.crt",
					Key:  "../example/var/lib/baetyl/testcert/server.key",
				},
			},
			cliConf: mockClientConf{
				Address: "https://127.0.0.1:50060",
				Certificate: utils.Certificate{
					Key:                "../example/var/lib/baetyl/testcert/client.key",
					Cert:               "../example/var/lib/baetyl/testcert/client.crt",
					InsecureSkipVerify: true,
				},
			},
		},
		{
			serverConf: ServerConfig{
				Address:     "127.0.0.1:50070",
				ReadTimeout: time.Second,
				Certificate: utils.Certificate{
					Cert:           "../example/var/lib/baetyl/testcert/server.crt",
					Key:            "../example/var/lib/baetyl/testcert/server.key",
					ClientAuthType: tls.VerifyClientCertIfGiven,
				},
			},
			cliConf: mockClientConf{
				Address: "https://127.0.0.1:50070",
				Certificate: utils.Certificate{
					CA:                 "../example/var/lib/baetyl/testcert/ca.crt",
					InsecureSkipVerify: true,
				},
			},
		},
	}

	for _, conf := range confs {
		server := NewServer(conf.serverConf, mockRoute())
		server.Start()
		time.Sleep(100 * time.Millisecond)

		client, err := newMockClient(conf.cliConf)
		assert.NoError(t, err)

		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		url := fmt.Sprintf("%s/%s", conf.cliConf.Address, "ping")
		req.SetRequestURI(url)
		req.Header.SetMethod("GET")
		err = client.Do(req, resp)
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode(), 200)
		data := map[string]string{}
		err = json.Unmarshal(resp.Body(), &data)
		assert.NoError(t, err)
		assert.Equal(t, "ok", data["status"])

		req2 := fasthttp.AcquireRequest()
		resp2 := fasthttp.AcquireResponse()
		url2 := fmt.Sprintf("%s/%s", conf.cliConf.Address, "err")
		req2.SetRequestURI(url2)
		req2.Header.SetMethod("GET")
		err2 := client.Do(req2, resp2)
		assert.NoError(t, err2)
		assert.Equal(t, resp2.StatusCode(), 500)

		req3 := fasthttp.AcquireRequest()
		resp3 := fasthttp.AcquireResponse()
		url3 := fmt.Sprintf("%s", conf.cliConf.Address)
		req3.SetRequestURI(url3)
		req3.Header.SetMethod("POST")
		err3 := client.Do(req3, resp3)
		assert.NoError(t, err3)
		assert.Equal(t, resp3.StatusCode(), 200)
		assert.Equal(t, string(resp3.Body()), "set")

		req4 := fasthttp.AcquireRequest()
		resp4 := fasthttp.AcquireResponse()
		url4 := fmt.Sprintf("%s/%s", conf.cliConf.Address, "any")
		req4.SetRequestURI(url4)
		req4.Header.SetMethod("DELETE")
		err4 := client.Do(req4, resp4)
		assert.NoError(t, err4)
		assert.Equal(t, resp4.StatusCode(), 200)
		assert.Equal(t, string(resp4.Body()), "delete")

		req5 := fasthttp.AcquireRequest()
		resp5 := fasthttp.AcquireResponse()
		url5 := fmt.Sprintf("%s/%s", conf.cliConf.Address, "any")
		req5.SetRequestURI(url5)
		req5.Header.SetMethod("PUT")
		err5 := client.Do(req5, resp5)
		assert.NoError(t, err5)
		assert.Equal(t, resp5.StatusCode(), 200)
		assert.Equal(t, string(resp5.Body()), "update")

		req6 := fasthttp.AcquireRequest()
		resp6 := fasthttp.AcquireResponse()
		url6 := fmt.Sprintf("%s/%s", conf.cliConf.Address, "stream")
		req6.SetRequestURI(url6)
		req6.Header.SetMethod("GET")
		err6 := client.Do(req6, resp6)
		assert.NoError(t, err6)
		assert.Equal(t, resp6.StatusCode(), 200)
		assert.Equal(t, string(resp6.Body()), "Hello, playground")
		server.Close()
	}
}

type mockClientConf struct {
	Address string
	utils.Certificate
}

func newMockClient(conf mockClientConf) (*fasthttp.Client, error) {
	_tls, err := utils.NewTLSConfigClient(conf.Certificate)
	if err != nil {
		return nil, err
	}
	client := &fasthttp.Client{
		TLSConfig: _tls,
	}
	return client, nil
}

func mockRoute() fasthttp.RequestHandler {
	router := routing.New()
	router.Get("/stream", mockStream)
	router.Get("/<key>", mockGet)
	router.Post("/", mockSet)
	router.Delete("/<key>", mockDelete)
	router.Put("/<key>", mockUpdate)

	return router.HandleRequest
}

// Get Get
func mockGet(c *routing.Context) error {
	key := c.Param("key")
	if key != "" && key != "err" {
		data := map[string]string{
			"status": "ok",
		}
		d, err := json.Marshal(data)
		if err != nil {
			RespondMsg(c, 500, "ERR_JSON", err.Error())
			return nil
		}
		Respond(c, http.StatusOK, d)
	} else {
		RespondMsg(c, 500, "ERR", "err")
	}
	return nil
}

// Set Set
func mockSet(c *routing.Context) error {
	Respond(c, http.StatusOK, []byte("set"))
	return nil
}

// Delete Delete
func mockDelete(c *routing.Context) error {
	Respond(c, http.StatusOK, []byte("delete"))
	return nil
}

// Update update
func mockUpdate(c *routing.Context) error {
	Respond(c, http.StatusOK, []byte("update"))
	return nil
}

// Stream Stream
func mockStream(c *routing.Context) error {
	a := "Hello, playground"
	reader := strings.NewReader(a)

	RespondStream(c, http.StatusOK, reader, -1)
	return nil
}
