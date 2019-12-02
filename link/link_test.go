package link

import (
	"context"
	"fmt"
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
				Name: "bd",
			},
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
		Content: []byte("timer ack"),
	}
)

func TestLink(t *testing.T) {
	wg := sync.WaitGroup{}
	sc := ServerConfig{
		Address: "0.0.0.0:8273",
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
	}
	sc.Message.Length.Max = 4194304
	s, err := NewServer(sc,
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
					msgResp.Context.ID = in.Context.ID
					msgResp.Context.TS = in.Context.TS
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
	assert.NoError(t, err)
	defer s.Close()

	handler := func(content []byte) []byte {
		assert.Equal(t, string(msgResp.Content), "video resp")
		return []byte("timer ack")
	}
	cc.Message.Length.Max = 4194304
	l, err := NewLinker(cc, handler)
	assert.NoError(t, err)
	defer l.Close()
	err = l.Send(msgSend.Context.Source, msgSend.Context.Destination, msgSend.Content)
	assert.NoError(t, err)
	err = l.stream.Send(packetAckMsg(msgResp, []byte("timer ack")))
	assert.NoError(t, err)
	wg.Add(1)
	wg.Wait()
}
