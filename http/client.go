package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	gohttp "net/http"
	"strings"
)

// Client client of http server
type Client struct {
	ops ClientOptions
	*gohttp.Client
}

// NewClient creates a new http client
func NewClient(ops ClientOptions) *Client {
	transport := &gohttp.Transport{
		Proxy: gohttp.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   ops.Timeout,
			KeepAlive: ops.KeepAlive,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          ops.MaxIdleConns,
		IdleConnTimeout:       ops.IdleConnTimeout,
		TLSHandshakeTimeout:   ops.TLSHandshakeTimeout,
		ExpectContinueTimeout: ops.ExpectContinueTimeout,
	}
	return &Client{
		ops: ops,
		Client: &gohttp.Client{
			Timeout:   ops.Timeout,
			Transport: transport,
		},
	}
}

// Call calls the function of service via HTTP POST
func (c *Client) Call(service, function string, payload []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.ops.Address, service)
	if function != "" {
		url += "/" + function
	}
	r, err := c.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	return HandleResponse(r)
}

// HandleResponse handles response
func HandleResponse(r *gohttp.Response) ([]byte, error) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if r.StatusCode != gohttp.StatusOK {
		msg := strings.TrimRight(string(data), "\n")
		if msg == "" {
			msg = r.Status
		}
		err = fmt.Errorf("[%d] %s", r.StatusCode, msg)
	}
	return data, err
}
