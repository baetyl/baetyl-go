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
		Context: Context{
			QOS:         1,
			Flags:       0,
			Topic:       "$SYS/service/video",
			Source:      "timer",
			Destination: "video",
		},
		Content: []byte("timer send"),
	}

	msgResp = &Message{
		Context: Context{
			ID:          123,
			TS:          3213123,
			QOS:         1,
			Flags:       0,
			Topic:       "$SYS/service/timer",
			Source:      "video",
			Destination: "timer",
		},
		Content: []byte("video send"),
	}

	msgAck = &Message{
		Context: Context{
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
		if in.Context.Flags != 4 {
			msgSend.Context.ID = in.Context.ID
			msgSend.Context.TS = in.Context.TS
			checkMsg(l.t, in, msgSend)
			if err = stream.Send(msgResp); err != nil {
				return err
			}
		} else {
			if err = stream.Send(packetAckMsg(in)); err != nil {
				return err
			}
			return nil
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

	ch := make(chan struct{})
	handler := func(m *Message) error {
		checkMsg(t, m, msgResp)
		close(ch)
		return nil
	}
	l, err := NewClient(cc, handler)
	assert.NoError(t, err)
	defer l.Close()
	err = l.Send(msgSend.Context.Source, msgSend.Context.Destination, 1, msgSend.Content)
	assert.NoError(t, err)
	<-ch
}
