package mock

import (
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl-go/v2/utils"
)

func TestHttpServer(t *testing.T) {
	tlssvr, err := utils.NewTLSConfigServer(utils.Certificate{CA: "./testcert/ca.pem", Key: "./testcert/server.key", Cert: "./testcert/server.pem"})
	assert.NoError(t, err)
	assert.NotNil(t, tlssvr)
	tlscli, err := utils.NewTLSConfigClient(utils.Certificate{CA: "./testcert/ca.pem", Key: "./testcert/client.key", Cert: "./testcert/client.pem", InsecureSkipVerify: true})
	assert.NoError(t, err)
	assert.NotNil(t, tlscli)

	svr := NewServer(tlssvr, NewResponse(200, []byte("hi")), NewResponse(400, nil))
	defer svr.Close()

	cli := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlscli,
			Proxy:           http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	res, err := cli.Post(svr.URL, "application/json", nil)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 200, res.StatusCode)
	data, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "hi", string(data))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		res, err := cli.Post(svr.URL, "application/json", nil)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, 400, res.StatusCode)
		data, err := ioutil.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "", string(data))
	}()
	wg.Wait()

	res, err = cli.Post(svr.URL, "application/json", nil)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 200, res.StatusCode)
	data, err = ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "", string(data))
}
