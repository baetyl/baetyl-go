package http

import (
	"fmt"
	"io/ioutil"
	gohttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Call(t *testing.T) {
	ts := httptest.NewServer(gohttp.HandlerFunc(func(w gohttp.ResponseWriter, r *gohttp.Request) {
		fmt.Printf("Header:%vn", r.Header)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "/service/function", r.URL.EscapedPath())
		req, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, "{}", string(req))
		w.WriteHeader(gohttp.StatusOK)
		w.Write(req)
	}))
	defer ts.Close()

	ops := NewClientOptions()
	ops.Address = ts.URL
	c := NewClient(ops)
	resp, err := c.Call("service", "function", []byte("{}"))
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(resp))
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
