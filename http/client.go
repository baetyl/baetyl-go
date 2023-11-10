package http

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	gohttp "net/http"
	"strings"
	"time"

	"github.com/conduitio/bwlimit"
	"github.com/panjf2000/ants/v2"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
)

var jsonHeaders = map[string]string{"Content-Type": "application/json"}

// Client client of http server
type Client struct {
	ops     *ClientOptions
	http    *gohttp.Client
	antPool *ants.Pool
}

// NewClient creates a new http client
func NewClient(ops *ClientOptions) *Client {
	transport := &gohttp.Transport{
		Proxy:                 gohttp.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		TLSClientConfig:       ops.TLSConfig,
		MaxIdleConns:          ops.MaxIdleConns,
		IdleConnTimeout:       ops.IdleConnTimeout,
		TLSHandshakeTimeout:   ops.TLSHandshakeTimeout,
		ExpectContinueTimeout: ops.ExpectContinueTimeout,
	}
	if ops.SpeedLimit != 0 {
		var speedLimit bwlimit.Byte
		speedLimit = bwlimit.Byte(ops.SpeedLimit)
		switch ops.ByteUnit {
		case ByteUnitMB:
			speedLimit = speedLimit * bwlimit.MB
		default:
			speedLimit = speedLimit * bwlimit.KB
		}
		bwlimitDialer := bwlimit.NewDialer(&net.Dialer{
			Timeout:   ops.Timeout,
			KeepAlive: ops.KeepAlive,
		}, 0, speedLimit)

		transport.DialContext = bwlimitDialer.DialContext
	} else {
		dialer := &net.Dialer{
			Timeout:   ops.Timeout,
			KeepAlive: ops.KeepAlive,
			DualStack: true,
		}
		transport.DialContext = dialer.DialContext
	}
	p, err := ants.NewPool(1)
	if err != nil {
		log.Error(errors.Errorf("http init pool error :%s", err))
	}
	if ops.SyncMaxConcurrency != 0 {
		p, err = ants.NewPool(ops.SyncMaxConcurrency)
		if err != nil {
			log.Error(errors.Errorf("http init pool error :%s", err))
		}
	}

	return &Client{
		ops: ops,
		http: &gohttp.Client{
			Timeout:   ops.Timeout,
			Transport: transport,
		},
		antPool: p,
	}
}

func (c *Client) SetBwlimit(writeLimit, readLimit bwlimit.Byte) {
	dialer := bwlimit.NewDialer(&net.Dialer{
		Timeout:   c.ops.Timeout,
		KeepAlive: c.ops.KeepAlive,
	}, writeLimit*bwlimit.Mebibyte, readLimit*bwlimit.KB)
	c.http.Transport.(*gohttp.Transport).DialContext = dialer.DialContext
}

// Call calls the function via HTTP POST
func (c *Client) Call(function string, payload []byte) ([]byte, error) {
	r, err := c.PostURL(function, bytes.NewBuffer(payload), jsonHeaders)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return HandleResponse(r)
}

// PostJSON post data with json content type
func (c *Client) PostJSON(url string, payload []byte, headers ...map[string]string) ([]byte, error) {
	headers = append(headers, jsonHeaders)
	r, err := c.PostURL(url, bytes.NewBuffer(payload), headers...)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return HandleResponse(r)
}

// GetJSON get data with json content type
func (c *Client) GetJSON(url string, headers ...map[string]string) ([]byte, error) {
	headers = append(headers, jsonHeaders)
	r, err := c.GetURL(url, headers...)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return HandleResponse(r)
}

func (c *Client) GetURL(url string, header ...map[string]string) (*gohttp.Response, error) {
	return c.SendUrl("GET", url, nil, header...)
}

func (c *Client) DeleteURL(url string, header ...map[string]string) (*gohttp.Response, error) {
	return c.SendUrl("DELETE", url, nil, header...)
}

func (c *Client) PostURL(url string, body io.Reader, header ...map[string]string) (*gohttp.Response, error) {
	return c.SendUrl("POST", url, body, header...)
}

func (c *Client) PutURL(url string, body io.Reader, header ...map[string]string) (*gohttp.Response, error) {
	return c.SendUrl("PUT", url, body, header...)
}

func (c *Client) SendUrl(method, url string, body io.Reader, header ...map[string]string) (*gohttp.Response, error) {
	if !strings.HasPrefix(url, "http") {
		url = fmt.Sprintf("%s/%s", c.ops.Address, url)
	}
	req, err := gohttp.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for _, v := range header {
		for kk, vv := range v {
			req.Header.Set(kk, vv)
		}
	}
	r, err := c.http.Do(req)
	return r, errors.Trace(err)
}

func (c *Client) SyncSendUrl(method, url string, body io.Reader, syncResult chan *SyncResults, extra map[string]interface{}, header ...map[string]string) {
	SyncSendStart := time.Now()
	err := c.antPool.Submit(
		func() {
			sendStart := time.Now()
			response, err := c.SendUrl(method, url, body, header...)
			sendElapsed := time.Since(sendStart)
			syncElapsed := time.Since(SyncSendStart)

			result := &SyncResults{
				Err:      err,
				Response: response,
				SendCost: sendElapsed,
				SyncCost: syncElapsed,
				Extra:    extra,
			}
			select {
			case syncResult <- result:
			default:
				log.Error(errors.New("can not add send result to syncResult from http con"))
			}
		})
	if err != nil {
		result := &SyncResults{
			Err:      err,
			Response: nil,
			SendCost: 0,
			SyncCost: 0,
			Extra:    extra,
		}
		select {
		case syncResult <- result:
		default:
			log.Error(errors.New("can not add send result to syncResult from http con"))
		}
	}
}

// HandleResponse handles response
func HandleResponse(r *gohttp.Response) ([]byte, error) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if r.StatusCode < gohttp.StatusOK || r.StatusCode > gohttp.StatusAlreadyReported {
		msg := strings.TrimRight(string(data), "\n")
		if msg == "" {
			msg = r.Status
		}
		err = errors.Errorf("[%d] %s", r.StatusCode, msg)
	}
	return data, errors.Trace(err)
}
