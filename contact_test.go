package baetyl

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/utils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestContactServer(t *testing.T) {
	scfg := ContactServerConfig{
		Address: "127.0.0.1:8080",
		Auth: Auth{
			Account: Account{
				Username: "test",
				Password: "test",
			},
			Certificate: utils.Certificate{
				Cert: "./testcert/server.pem",
				Key:  "./testcert/server.key",
			},
		},
		Timeout: time.Duration(30) * time.Second,
	}
	ser, err := NewCServer(scfg, func(ctx context.Context, msg *Message, option ...grpc.CallOption) (message *Message, e error) {
		return msg, nil
	}, func(stream Contact_TalkServer) error {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			stream.Send(in)
		}
	})
	assert.NoError(t, err)
	defer ser.Close()
}

func TestContactClient(t *testing.T) {
	t.Skip("local test: start server")
	ccfg := ContactClientConfig{
		Address: "127.0.0.1:8080",
		Auth: Auth{
			Account: Account{
				Username: "test",
				Password: "test",
			},
			Certificate: utils.Certificate{
				Cert: "./testcert/server.pem",
				Key:  "./testcert/server.key",
			},
		},
		Timeout: time.Duration(30) * time.Second,
	}
	ccfgErr0 := ContactClientConfig{
		Address: "127.0.0.1:8080",
		Auth: Auth{
			Account: Account{
				Username: "test",
				Password: "test",
			},
			Certificate: utils.Certificate{
				Cert: "./testcert/server.pem",
				Key:  "",
			},
		},
		Timeout: time.Duration(30) * time.Second,
	}
	ccfgErr1 := ContactClientConfig{
		Address: "127.0.0.1:8080",
		Auth: Auth{
			Account: Account{
				Username: "test",
				Password: "1234",
			},
			Certificate: utils.Certificate{
				Cert: "./testcert/server.pem",
				Key:  "./testcert/server.key",
			},
		},
		Timeout: time.Duration(30) * time.Second,
	}

	msg := &Message{
		Context: &MsgContext{
			ID:     1,
			TS:     123456,
			QOS:    2,
			Retain: 77,
			Topic:  "baety/grpc/cli/ser/test",
			Src:    "cli",
			Dest:   "ser",
		},
		Content: []byte("test123"),
	}

	cli, err := NewCClient(ccfg)
	assert.NoError(t, err)
	defer cli.Close()
	start := time.Now()
	resp, err := cli.Call(msg)
	fmt.Printf("%s elapsed time: %v\n", t.Name(), time.Since(start))
	assert.NoError(t, err)
	assert.Equal(t, "test123", string(resp.Content))
	assert.Equal(t, 1, resp.Context.ID)
	assert.Equal(t, 123456, resp.Context.TS)
	assert.Equal(t, 2, resp.Context.QOS)
	assert.Equal(t, 77, resp.Context.Retain)
	assert.Equal(t, "baety/grpc/cli/ser/test", resp.Context.Topic)
	assert.Equal(t, "cli", resp.Context.Src)
	assert.Equal(t, "ser", resp.Context.Dest)

	cliErr, err := NewCClient(ccfgErr0)
	assert.Error(t, err)
	cliErr, err = NewCClient(ccfgErr1)
	assert.NoError(t, err)
	defer cliErr.Close()
	_, err = cliErr.Call(msg)
	assert.Error(t, err)
}
