package link

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLinkClient(t *testing.T) {
	t.Skip("local svr: start server")
	ccfg := LClientConfig{
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
	ccfgErr0 := LClientConfig{
		Address: "127.0.0.1:8080",
		Timeout: time.Duration(1) * time.Second,
		Account: Account{
			Username: "svr",
			Password: "svr",
		},
		Certificate: LClientCert{
			Cert: "./testcert/server.pem",
			Name: "",
		},
	}
	ccfgErr1 := LClientConfig{
		Address: "127.0.0.1:8080",
		Timeout: time.Duration(20) * time.Second,
		Account: Account{
			Username: "svr",
			Password: "error",
		},
		Certificate: LClientCert{
			Cert: "./testcert/server.pem",
			Name: "bd",
		},
	}

	msg := &Message{
		Context: &Context{
			ID:    1,
			TS:    123456,
			QOS:   2,
			Flags: 77,
			Topic: "baety/grpc/cli/ser/svr",
			Src:   "cli",
			Dest:  "ser",
		},
		Content: []byte("test123"),
	}

	// cert error
	_, err := NewLClient(ccfgErr0)
	assert.Error(t, err)

	// username & password error
	cliErr, err := NewLClient(ccfgErr1)
	assert.NoError(t, err)
	defer cliErr.Close()
	_, err = cliErr.Call(msg)
	assert.Error(t, err)
	assert.Equal(t, "rpc error: code = Unauthenticated "+
		"desc = username or password not match", err.Error())

	// Call
	cli, err := NewLClient(ccfg)
	assert.NoError(t, err)
	defer cli.Close()
	start := time.Now()
	resp, err := cli.Call(msg)
	fmt.Printf("%s call elapsed time: %v\n", t.Name(), time.Since(start))
	assert.NoError(t, err)
	assert.Equal(t, string(msg.Content), string(resp.Content))
	assert.Equal(t, msg.Context.ID, resp.Context.ID)
	assert.Equal(t, msg.Context.TS, resp.Context.TS)
	assert.Equal(t, msg.Context.QOS, resp.Context.QOS)
	assert.Equal(t, msg.Context.Flags, resp.Context.Flags)
	assert.Equal(t, msg.Context.Topic, resp.Context.Topic)
	assert.Equal(t, msg.Context.Src, resp.Context.Src)
	assert.Equal(t, msg.Context.Dest, resp.Context.Dest)

	// Talk
	msg = &Message{
		Context: &Context{
			ID:    2,
			TS:    654321,
			QOS:   1,
			Flags: 44,
			Topic: "baety/grpc/cli/ser/cli",
			Src:   "timer",
			Dest:  "python",
		},
		Content: []byte("test for talk"),
	}
	waitc := make(chan struct{})
	stream, err := cli.Talk()
	assert.NoError(t, err)
	go func() {
		in, err := stream.Recv()
		assert.NoError(t, err)
		resp = in
		close(waitc)
	}()
	err = stream.Send(msg)
	assert.NoError(t, err)
	err = stream.CloseSend()
	assert.NoError(t, err)
	<-waitc

	assert.Equal(t, string(msg.Content), string(resp.Content))
	assert.Equal(t, msg.Context.ID, resp.Context.ID)
	assert.Equal(t, msg.Context.TS, resp.Context.TS)
	assert.Equal(t, msg.Context.QOS, resp.Context.QOS)
	assert.Equal(t, msg.Context.Flags, resp.Context.Flags)
	assert.Equal(t, msg.Context.Topic, resp.Context.Topic)
	assert.Equal(t, msg.Context.Src, resp.Context.Src)
	assert.Equal(t, msg.Context.Dest, resp.Context.Dest)
}
