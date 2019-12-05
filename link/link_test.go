package link

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/utils"

	"github.com/stretchr/testify/assert"
)

var (
	cc = ClientConfig{
		Address: "0.0.0.0:8273",
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
			},
		},
		MaxSize: 464471,
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

	msgResp = &Message{
		Context: &Context{
			QOS:         1,
			Flags:       4,
			Topic:       "$SYS/service/video",
			Source:      "video",
			Destination: "timer",
		},
		Content: []byte("video resp"),
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
	t  *testing.T
	wg sync.WaitGroup
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
		if string(in.Content) != "timer ack" {
			msgSend.Context.ID = in.Context.ID
			msgSend.Context.TS = in.Context.TS
			checkMsg(l.t, in, msgSend)
			msgResp.Context.ID = in.Context.ID
			msgResp.Context.TS = in.Context.TS
			if err = stream.Send(msgResp); err != nil {
				return err
			}
			continue
		}
		msgAck.Context.ID = in.Context.ID
		msgAck.Context.TS = in.Context.TS
		checkMsg(l.t, in, msgAck)
		l.wg.Done()
		break
	}
	return nil
}

func TestLink(t *testing.T) {
	wg := sync.WaitGroup{}
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
		MaxSize: 464471,
	}
	svr, err := NewServer(sc)
	assert.NoError(t, err)
	defer svr.GracefulStop()
	s := &lkSerLT{
		t:  t,
		wg: wg,
	}
	RegisterLinkServer(svr, s)
	lis, err := net.Listen("tcp", "0.0.0.0:8273")
	assert.NoError(t, err)
	go svr.Serve(lis)

	handler := func(m *Message) error {
		assert.Equal(t, string(msgResp.Content), "video resp")
		return nil
	}
	l, err := NewClient(cc, handler)
	assert.NoError(t, err)
	defer l.Close()
	err = l.Send(msgSend.Context.Source, msgSend.Context.Destination, 1, msgSend.Content)
	assert.NoError(t, err)
	err = l.stream.Send(packetAckMsg(msgResp))
	assert.NoError(t, err)
	wg.Add(1)
	wg.Wait()
}
