package link

import (
	"context"
	fmt "fmt"
	"sync"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/mock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type mockObserver struct {
	t        *testing.T
	msgs     chan *Message
	errs     chan error
	errOnMsg error
	sync.Mutex
}

func newMockObserver(t *testing.T) *mockObserver {
	return &mockObserver{
		t:    t,
		msgs: make(chan *Message, 10),
		errs: make(chan error, 10),
	}
}

func (o *mockObserver) OnMsg(msg *Message) error {
	fmt.Printf("--> OnMsg: %v <--\n", msg)
	select {
	case o.msgs <- msg:
	default:
	}
	o.Lock()
	defer o.Unlock()
	return o.errOnMsg
}

func (o *mockObserver) OnAck(msg *Message) error {
	fmt.Printf("--> OnAck: %v <--\n", msg)
	select {
	case o.msgs <- msg:
	default:
	}
	return nil
}

func (o *mockObserver) OnErr(err error) {
	fmt.Printf("--> OnErr: %v <--\n", err)
	select {
	case o.errs <- err:
	default:
	}
}

func (o *mockObserver) setErrOnMsg(err error) {
	o.Lock()
	o.errOnMsg = err
	o.Unlock()
}

func (o *mockObserver) assertMsgs(msgs ...*Message) {
	for _, msg := range msgs {
		select {
		case <-time.After(1 * time.Minute):
			panic("nothing received")
		case p := <-o.msgs:
			assert.Equal(o.t, msg, p)
		}
	}
}

func (o *mockObserver) assertErrs(errs ...error) {
	for _, err := range errs {
		select {
		case <-time.After(1 * time.Second):
			panic("nothing received")
		case e := <-o.errs:
			assert.Equal(o.t, err.Error(), e.Error())
		}
	}
}

func safeReceive(ch chan struct{}) {
	select {
	case <-time.After(1 * time.Second):
		panic("nothing received")
	case <-ch:
	}
}

func newClientOptions(t *testing.T) ClientOptions {
	o := NewClientOptions()
	o.Address = "0.0.0.0:50006"
	o.Observer = newMockObserver(t)
	return o
}

type mockServer struct {
	t *testing.T
	s *grpc.Server
	f *mock.Flow
	q chan struct{}
	sync.Once
}

func (s *mockServer) Call(ctx context.Context, msg *Message) (*Message, error) {
	return msg, nil
}

func (s *mockServer) Talk(stream Link_TalkServer) error {
	fmt.Println("server starts to talk")
	defer fmt.Println("server has stopped talking")

	err := s.f.Test(newWrapper(s, stream))
	fmt.Println("server test error:", err)
	assert.NoError(s.t, err)
	return nil
}

func (s *mockServer) Close() error {
	s.Do(func() {
		fmt.Println("server stops")
		defer fmt.Println("server has stopped")
		s.s.Stop()
		close(s.q)
	})
	return nil
}

// initMockServer the fake of link server for test only
func initMockServer(t *testing.T, f *mock.Flow) chan struct{} {
	ms := &mockServer{t: t, f: f, q: make(chan struct{})}

	ops := NewServerOptions()
	ops.Address = "link://0.0.0.0:50006"
	ops.LinkServer = ms
	svr, err := Launch(ops)
	assert.NoError(t, err)

	ms.s = svr
	return ms.q
}

type wrapper struct {
	server *mockServer
	stream Link_TalkServer
}

func newWrapper(s *mockServer, conn Link_TalkServer) mock.Conn {
	return &wrapper{server: s, stream: conn}
}

func (c *wrapper) Send(msg interface{}) error {
	return c.stream.Send(msg.(*Message))
}

func (c *wrapper) Receive() (interface{}, error) {
	msg, err := c.stream.Recv()
	fmt.Println("server stream received:", msg, err)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *wrapper) Close() error {
	return c.server.Close()
}
