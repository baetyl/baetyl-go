package http

import (
	"encoding/json"
	routing "github.com/qiangxue/fasthttp-routing"
)

const (
	jsonContentTypeHeader = "application/json"
)

// Response Response
type Response struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

// NewResponse NewResponse
func NewResponse(code, msg string) Response {
	return Response{
		Code: code,
		Msg:  msg,
	}
}

func RespondMsg(c *routing.Context, httpCode int, code, msg string) {
	resp := NewResponse(code, msg)
	b, _ := json.Marshal(&resp)
	Respond(c, httpCode, b)
}

func Respond(c *routing.Context, httpCode int, obj []byte) {
	c.RequestCtx.Response.SetStatusCode(httpCode)
	c.RequestCtx.Response.SetBody(obj)
	if json.Valid(obj) {
		c.RequestCtx.Response.Header.SetContentType(jsonContentTypeHeader)
	}
}
