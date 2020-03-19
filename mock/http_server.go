package mock

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
)

// Response the mocked http repsonse
type Response struct {
	status int
	body   []byte
}

// NewResponse create a new repsonse
func NewResponse(status int, body []byte) *Response {
	return &Response{status, body}
}

// Server the mocked http server
type Server struct {
	*httptest.Server
	tlsConfig *tls.Config
	responses []*Response
}

// NewServer create a new mocked server
func NewServer(tlsConfig *tls.Config, responses ...*Response) *Server {
	ms := &Server{tlsConfig: tlsConfig, responses: responses}
	ms.Server = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(ms.responses) == 0 {
			return
		}
		w.WriteHeader(ms.responses[0].status)
		w.Write(ms.responses[0].body)
		ms.responses = ms.responses[1:]
	}))
	if tlsConfig == nil {
		ms.Server.Start()
	} else {
		ms.Server.Config.TLSConfig = tlsConfig
		ms.Server.StartTLS()
	}
	return ms
}
