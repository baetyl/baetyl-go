package http

import (
	"encoding/json"
	"io"

	routing "github.com/qiangxue/fasthttp-routing"
)

const (
	jsonContentTypeHeader = "application/json"
)

// Response Response
type Response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewResponse NewResponse
func NewResponse(code, msg string) Response {
	return Response{
		Code:    code,
		Message: msg,
	}
}

// RespondMsg RespondMsg
func RespondMsg(c *routing.Context, httpCode int, code, msg string) {
	resp := NewResponse(code, msg)
	b, _ := json.Marshal(&resp)
	Respond(c, httpCode, b)
}

// Respond Respond
func Respond(c *routing.Context, httpCode int, obj []byte) {
	c.RequestCtx.Response.SetStatusCode(httpCode)
	c.RequestCtx.Response.SetBody(obj)
	if json.Valid(obj) {
		c.RequestCtx.Response.Header.SetContentType(jsonContentTypeHeader)
	}
}

// RespondStream RespondStream
// If bodySize < 0, then bodyStream is read until io.EOF.
func RespondStream(c *routing.Context, httpCode int, bodyStream io.Reader, bodySize int) {
	c.RequestCtx.Response.SetStatusCode(httpCode)
	c.RequestCtx.Response.SetBodyStream(bodyStream, bodySize)
}
