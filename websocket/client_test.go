package websocket

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func Test_client(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 升级连接为websocket协议
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade HTTP connection to WebSocket: %v", err)
		}
		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Failed to read message from WebSocket: %v", err)
		}
		err = conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			t.Fatalf("Failed to write message to WebSocket: %v", err)
		}

	}))
	defer server.Close()

	cfg := ClientConfig{
		Address:             server.URL[len("http://"):],
		Path:                "",
		Schema:              "ws",
		IdleConnTimeout:     0,
		TLSHandshakeTimeout: 0,
		SyncMaxConcurrency:  10,
		Certificate:         utils.Certificate{},
	}
	options, err := cfg.ToClientOptions()
	assert.NoError(t, err)

	var msg []chan *ReadMsg
	for i := 0; i < options.SyncMaxConcurrency; i++ {
		msg = append(msg, make(chan *ReadMsg, 1))
	}

	client, err := NewClient(options, msg)
	assert.NoError(t, err)
	result := make(chan *SyncResults, 100)
	extra := map[string]interface{}{"a": 1}

	for i := 0; i < 100; i++ {
		client.SyncSendMsg([]byte("hello"), result, extra)
	}
	time.Sleep(time.Second)

	re := <-result
	assert.NoError(t, re.Err)
	assert.Equal(t, re.Extra["a"], 1)
	assert.Equal(t, 99, len(result))

	for _, m := range msg {
		r := <-m
		assert.Equal(t, r.Data, []byte("hello"))
		assert.NoError(t, r.Err)
	}

}
