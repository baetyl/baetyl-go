package link

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	g "github.com/baetyl/baetyl-go/utils/protocol/grpc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

var (
	cc = ClientConfig{
		Address: "0.0.0.0:8273",
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
			ID:          1,
			TS:          2,
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
			Flags:       0,
			Topic:       "$SYS/service/video",
			Source:      "timer",
			Destination: "video",
		},
		Content: []byte("timer ack"),
	}
)

func TestLink(t *testing.T) {
	wg := sync.WaitGroup{}
	option := &g.ServerOption{}
	opts := option.
		Create().
		CredsFromFile("./testcert/server.pem", "./testcert/server.key").
		Build()
	s := NewServer("svr",
		"svr",
		func(c context.Context, msg *Message) (*Message, error) {
			checkMsg(t, msg, msgCall)
			return msgCallResp, nil
		},
		func(stream Link_TalkServer) error {
			for {
				in, err := stream.Recv()
				if err != nil {
					return err
				}
				fmt.Printf("server receive = %v\n", in)
				if string(in.Content) != "timer ack" {
					msgSend.Context.ID = in.Context.ID
					msgSend.Context.TS = in.Context.TS
					checkMsg(t, in, msgSend)
					if err = stream.Send(msgResp); err != nil {
						return err
					}
					continue
				}
				msgAck.Context.ID = in.Context.ID
				msgAck.Context.TS = in.Context.TS
				checkMsg(t, in, msgAck)
				wg.Done()
				break
			}
			return nil
		})
	svr, err := g.NewServer(g.NetTCP, "0.0.0.0:8273", opts, func(svr *grpc.Server) {
		RegisterLinkServer(svr, s)
	})
	assert.NoError(t, err)
	defer svr.GracefulStop()

	handler := func(content []byte) []byte {
		assert.Equal(t, string(msgResp.Content), "video resp")
		return []byte("timer ack")
	}
	l, err := NewLinker(cc, handler)
	assert.NoError(t, err)
	err = l.Send(msgSend.Context.Source, msgSend.Context.Destination, msgSend.Content, 10000)
	assert.NoError(t, err)
	wg.Add(1)
	wg.Wait()
	l.Close()
}
