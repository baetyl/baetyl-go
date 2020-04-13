package http

import (
	"encoding/json"
	routing "github.com/qiangxue/fasthttp-routing"
)

const (
	jsonContentTypeHeader = "application/json"
)

// ErrorResponse ErrorResponse
type ErrorResponse struct {
	ErrCode string `json:"errCode"`
	Message string `json:"message"`
}

// NewErrorResponse NewErrorResponse
func NewErrorResponse(errCode, message string) ErrorResponse {
	return ErrorResponse{
		ErrCode: errCode,
		Message: message,
	}
}

func respondError(c *routing.Context, code int, errCode, msg string) {
	resp := NewErrorResponse(errCode, msg)
	b, _ := json.Marshal(&resp)
	respond(c, code, b)
}

func respond(c *routing.Context, code int, obj []byte) {
	c.RequestCtx.Response.SetStatusCode(code)
	c.RequestCtx.Response.SetBody(obj)
	if json.Valid(obj) {
		c.RequestCtx.Response.Header.SetContentType(jsonContentTypeHeader)
	}
}
