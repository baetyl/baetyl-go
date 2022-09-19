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

type RPCMqttMessage struct {
	QoS     uint32      `json:"qos,omitempty"`
	Topic   string      `json:"topic,omitempty"`
	Content interface{} `json:"content,omitempty"`
}
