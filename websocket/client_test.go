package websocket

import (
	"log"
	"net/http"
	"testing"
	"time"

	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func echo(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		err = c.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("read:", err)
		}

	}
}
func WsServer() {
	http.HandleFunc("/echo", echo)
	log.Fatal(http.ListenAndServe("127.0.0.1:9341", nil))
}
func Test_client(t *testing.T) {
	go WsServer()

	cfg := ClientConfig{
		Address:             "127.0.0.1:9341",
		Path:                "echo",
		Schema:              "ws",
		IdleConnTimeout:     0,
		TLSHandshakeTimeout: 0,
		SyncMaxConcurrency:  10,
		Certificate:         utils.Certificate{},
	}
	options, err := cfg.ToClientOptions()
	assert.NoError(t, err)

	var msg []chan *v1.Message
	for i := 0; i < options.SyncMaxConcurrency; i++ {
		msg = append(msg, make(chan *v1.Message, 1))
	}

	client, err := NewClient(options, msg)
	assert.NoError(t, err)
	result := make(chan *SyncResults, 1000)
	extra := map[string]interface{}{"a": 1}

	time.Sleep(time.Second * 2)
	for i := 0; i < 20; i++ {
		client.SyncSendMsg([]byte("hello"), result, extra)
	}
	time.Sleep(time.Second * 2)
	re := <-result
	assert.NoError(t, re.Err)
	assert.Equal(t, re.Extra["a"], 1)
	assert.Equal(t, 19, len(result))

	for _, m := range msg {
		r := <-m
		assert.Equal(t, r.Content.Value.(WebsocketReadMsg).Data, []byte("hello"))
		assert.Equal(t, r.Kind, v1.MessageWebsocketRead)
	}

}
