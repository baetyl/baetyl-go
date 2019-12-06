package link

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/utils"

	"github.com/stretchr/testify/assert"
)

var (
	cc = ClientConfig{
		Address: "0.0.0.0:8274",
		Timeout: time.Duration(20) * time.Second,
		Auth: Auth{
			Account: Account{
				Username: "svr",
				Password: "svr",
			},
			Certificate: utils.Certificate{
				Cert: "./testcert/client.pem",
				Key:  "./testcert/client.key",
				CA:   "./testcert/ca.pem",
				Name: "bd",
			},
		},
		MaxMessageSize: 464471,
		DisableAutoAck: false,
	}

	msgSend = &Message{
		Context: &Context{
			QOS:         1,
			Flags:       0,
			Topic:       "$SYS/service/video",
			Source:      "timer",
			Destination: "video",
		},
		Content: []byte("timer send"),
	}

	msgAck = &Message{
		Context: &Context{
			QOS:         1,
			Flags:       4,
			Topic:       "$SYS/service/video",
			Source:      "timer",
			Destination: "video",
		},
		Content: nil,
	}
)

type lkSerLT struct {
	t *testing.T
}

func (l *lkSerLT) Call(ctx context.Context, msg *Message) (*Message, error) {
	checkMsg(l.t, msg, msgCall)
	return msgCallResp, nil
}

func (l *lkSerLT) Talk(stream Link_TalkServer) error {
	for {
		in, err := stream.Recv()
		if err != nil {
			return err
		}
		fmt.Printf("server receive = %v\n", in)
		msgSend.Context.ID = in.Context.ID
		msgSend.Context.TS = in.Context.TS
		checkMsg(l.t, in, msgSend)
		msgAck.Context.ID = in.Context.ID
		msgAck.Context.TS = in.Context.TS
		if err = stream.Send(msgAck); err != nil {
			return err
		}
	}
}

func TestLink(t *testing.T) {
	sc := ServerConfig{
		Auth: Auth{
			Account: Account{
				Username: "svr",
				Password: "svr",
			},
			Certificate: utils.Certificate{
				Cert: "./testcert/server.pem",
				Key:  "./testcert/server.key",
				CA:   "./testcert/ca.pem",
			},
		},
		MaxMessageSize: 464471,
	}
	svr, err := NewServer(sc)
	assert.NoError(t, err)
	s := &lkSerLT{
		t: t,
	}
	RegisterLinkServer(svr, s)
	lis, err := net.Listen("tcp", "0.0.0.0:8274")
	assert.NoError(t, err)
	go svr.Serve(lis)

	handler := func(m *Message) error {
		return nil
	}
	l, err := NewClient(cc, handler)
	assert.NoError(t, err)
	defer l.Close()
	err = l.Send(msgSend.Context.Source, msgSend.Context.Destination, 1, msgSend.Content)
	assert.NoError(t, err)
}
