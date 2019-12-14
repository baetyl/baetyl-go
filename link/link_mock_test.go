package link

import (
	"context"
	fmt "fmt"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/flow"
	"github.com/creasty/defaults"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var testAddr = "127.0.0.1:50006"

type mockObserver struct {
	t    *testing.T
	msgs chan *Message
	errs chan error
}

func newMockObserver(t *testing.T) *mockObserver {
	return &mockObserver{
		t:    t,
		msgs: make(chan *Message, 10),
		errs: make(chan error, 10),
	}
}

func (o *mockObserver) OnMsg(msg *Message) error {
	o.msgs <- msg
	return nil
}

func (o *mockObserver) OnAck(msg *Message) error {
	o.msgs <- msg
	return nil
}

func (o *mockObserver) OnErr(err error) {
	fmt.Printf("--> OnErr: %s <--\n", err.Error())
	o.errs <- err
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

func newServerConfig() (c ServerConfig) {
	c.Address = testAddr
	defaults.Set(&c)
	return
}

func newClientConfig() (c ClientConfig) {
	c.Address = testAddr
	c.Username = "u1"
	c.Password = "p1"
	c.DisableAutoAck = true
	defaults.Set(&c)
	return
}

type mockAuth map[string]string

func (ma mockAuth) Authenticate(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ErrUnauthenticated
	}
	u, ok := md[KeyUsername]
	if !ok || len(u) != 1 {
		return ErrUnauthenticated
	}
	p, ok := md[KeyPassword]
	if !ok || len(p) != 1 {
		return ErrUnauthenticated
	}
	if v, ok := ma[u[0]]; !ok || v != p[0] {
		return ErrUnauthenticated
	}
	return nil
}

type mockServer struct {
	t *testing.T
	s *grpc.Server
	f *flow.Flow
	q chan struct{}
	sync.Once
}

func (s *mockServer) Call(ctx context.Context, msg *Message) (*Message, error) {
	return msg, nil
}

func (s *mockServer) Talk(stream Link_TalkServer) error {
	fmt.Println("server starts to talk")
	defer fmt.Println("server has stopped talking")

	err := s.f.Test(newWrapper(stream))
	assert.NoError(s.t, err)

	s.Do(func() {
		close(s.q)
	})
	return nil
}

func (s *mockServer) Close() {
	s.s.Stop()
}

func fakeServer(t *testing.T, f *flow.Flow) *mockServer {
	ma := mockAuth{
		"u1": "p1",
		"u2": "p2",
	}
	s, err := NewServer(newServerConfig(), ma)
	assert.NoError(t, err)

	ms := &mockServer{t: t, s: s, f: f, q: make(chan struct{})}
	RegisterLinkServer(s, ms)

	lis, err := net.Listen("tcp", testAddr)
	assert.NoError(t, err)
	go s.Serve(lis)
	return ms
}

type wrapper struct {
	stream Link_TalkServer
}

func newWrapper(conn Link_TalkServer) flow.Conn {
	return &wrapper{stream: conn}
}

func (c *wrapper) Send(msg interface{}) error {
	return c.stream.Send(msg.(*Message))
}

func (c *wrapper) Receive() (interface{}, error) {
	msg, err := c.stream.Recv()
	if err != nil && strings.Contains(err.Error(), "Canceled") {
		return nil, io.EOF
	}
	return msg, nil
}

func (c *wrapper) Close() error {
	return nil
}
