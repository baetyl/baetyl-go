package link

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLinkSendReceive(t *testing.T) {
	cfg := LClientConfig{
		Address: "127.0.0.1:8080",
		Timeout: time.Duration(20) * time.Second,
		Account: Account{
			Username: "svr",
			Password: "svr",
		},
		Certificate: LClientCert{
			Cert: "./testcert/server.pem",
			Name: "bd",
		},
	}

	lTimer := NewLinker(cfg)
	defer lTimer.Close()
	err := lTimer.Send("timer", "video", []byte("test timer"))
	assert.NoError(t, err)

	handler := func(resp *Message) (bytes []byte, e error) {
		fmt.Printf("Receive1 %v\n", resp)
		assert.Equal(t, "test video", string(resp.Content))
		assert.Equal(t, uint32(1), resp.Context.QOS)
		assert.Equal(t, uint32(0), resp.Context.Flags)
		assert.Equal(t, "$sys/service/timer", resp.Context.Topic)
		assert.Equal(t, "video", resp.Context.Src)
		assert.Equal(t, "timer", resp.Context.Dest)
		return nil, nil
	}
	lTimer.Receive(handler)
	lTimer.Receive(nil)
	lTimer.Receive(handler)
	lTimer.Receive(handler)
	lTimer.Receive(nil)
	err = lTimer.Send("timer", "video", []byte("test timer"))
	if err != nil {
		fmt.Println(err.Error())
	}
	err = lTimer.Send("timer", "video", []byte("test timer"))
	if err != nil {
		fmt.Println(err.Error())
	}
	lTimer.Receive(nil)
	lTimer.Receive(nil)
	lTimer.Receive(handler)
	err = lTimer.Send("timer", "video", []byte("test timer"))
	if err != nil {
		fmt.Println(err.Error())
	}
	lTimer.Receive(handler)
	lTimer.Receive(handler)

	time.Sleep(time.Duration(50) * time.Millisecond)
	waitc := make(chan struct{})
	err = lTimer.Send("timer", "video", []byte("test timer"))
	if err != nil {
		fmt.Println(err.Error())
	}
	lTimer.Receive(nil)
	lTimer.Receive(nil)
	time.Sleep(time.Duration(50) * time.Millisecond)
	lTimer.Receive(func(resp *Message) (bytes []byte, e error) {
		fmt.Printf("Receive2 %v\n", resp)
		assert.Equal(t, "test video", string(resp.Content))
		assert.Equal(t, uint32(1), resp.Context.QOS)
		assert.Equal(t, uint32(0), resp.Context.Flags)
		assert.Equal(t, "$sys/service/timer", resp.Context.Topic)
		assert.Equal(t, "video", resp.Context.Src)
		assert.Equal(t, "timer", resp.Context.Dest)
		close(waitc)
		return nil, nil
	})
	<-waitc
}

func TestLinkSync(t *testing.T) {
	cfg := LClientConfig{
		Address: "127.0.0.1:8080",
		Timeout: time.Duration(20) * time.Second,
		Account: Account{
			Username: "svr",
			Password: "svr",
		},
		Certificate: LClientCert{
			Cert: "./testcert/server.pem",
			Name: "bd",
		},
	}
	lTimer := NewLinker(cfg)
	resp, err := lTimer.SendSync("timer", "video", []byte("test timer"), time.Duration(30)*time.Second)
	assert.NoError(t, err)
	fmt.Printf("resp %v\n", resp)
	assert.Equal(t, "test video", string(resp.Content))
	assert.Equal(t, uint32(1), resp.Context.QOS)
	assert.Equal(t, uint32(4), resp.Context.Flags)
	assert.Equal(t, "$sys/service/timer", resp.Context.Topic)
	assert.Equal(t, "video", resp.Context.Src)
	assert.Equal(t, "timer", resp.Context.Dest)
	resp, err = lTimer.SendSync("timer", "video", []byte("test timer"), time.Duration(1)*time.Second)
	assert.Error(t, err)
	assert.Equal(t, "timeout", err.Error())
}
