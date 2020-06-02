package http

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	gohttp "net/http"
	"strings"
)

// ContentTypeJSON the json content type of request
const ContentTypeJSON = "application/json"

// Client client of http server
type Client struct {
	ops  *ClientOptions
	http *gohttp.Client
}

// NewClient creates a new http client
func NewClient(ops *ClientOptions) *Client {
	transport := &gohttp.Transport{
		Proxy: gohttp.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   ops.Timeout,
			KeepAlive: ops.KeepAlive,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		TLSClientConfig:       ops.TLSConfig,
		MaxIdleConns:          ops.MaxIdleConns,
		IdleConnTimeout:       ops.IdleConnTimeout,
		TLSHandshakeTimeout:   ops.TLSHandshakeTimeout,
		ExpectContinueTimeout: ops.ExpectContinueTimeout,
	}
	return &Client{
		ops: ops,
		http: &gohttp.Client{
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
	r, err := c.http.Post(url, ContentTypeJSON, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	return HandleResponse(r)
}

// PostJSON post data with json content type
func (c *Client) PostJSON(url string, payload []byte) ([]byte, error) {
	url = fmt.Sprintf("%s/%s", c.ops.Address, url)
	r, err := c.http.Post(url, ContentTypeJSON, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	return HandleResponse(r)
}

// GetJSON get data with json content type
func (c *Client) GetJSON(url string) ([]byte, error) {
	url = fmt.Sprintf("%s/%s", c.ops.Address, url)
	r, err := c.http.Get(url)
	if err != nil {
		return nil, err
	}
	return HandleResponse(r)
}

func (c *Client) GetURL(url string, header ...map[string]string) (*gohttp.Response, error) {
	req, err := gohttp.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for _, v := range header {
		for kk, vv := range v {
			req.Header.Set(kk, vv)
		}
	}
	return c.http.Do(req)
}

func (c *Client) PostURL(url string, body io.Reader, header ...map[string]string) (*gohttp.Response, error) {
	req, err := gohttp.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	for _, v := range header {
		for kk, vv := range v {
			req.Header.Set(kk, vv)
		}
	}
	return c.http.Do(req)
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
