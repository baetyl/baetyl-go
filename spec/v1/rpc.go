package v1

type RPCRequest struct {
	App    string            `json:"app"`
	Method string            `json:"method"`
	System bool              `json:"system,omitempty" default:"false"`
	Params string            `json:"params,omitempty"`
	Header map[string]string `json:"header,omitempty"`
	Body   interface{}       `json:"body,omitempty"`
}

type RPCResponse struct {
	StatusCode int                 `json:"statusCode"`
	Header     map[string][]string `json:"header,omitempty"`
	Body       []byte              `json:"body,omitempty"`
}
