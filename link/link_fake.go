package link

import (
	"context"
	fmt "fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/flow"
	"github.com/creasty/defaults"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var testAddr = "0.0.0.0:50006"

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
	fmt.Printf("--> OnMsg: %v <--\n", msg)
	select {
	case o.msgs <- msg:
	default:
	}
	return nil
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
	defaults.Set(&c)
	return
}

// FakeAuth the fake of auth
type FakeAuth map[string]string

// Authenticate authenticates username and password
func (ma FakeAuth) Authenticate(ctx context.Context) error {
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

// FakeServer the fake of link server for test only
func FakeServer(t *testing.T, f *flow.Flow, a Authenticator) chan struct{} {
	s, err := NewServer(newServerConfig(), a)
	assert.NoError(t, err)

	ms := &mockServer{t: t, s: s, f: f, q: make(chan struct{})}
	RegisterLinkServer(s, ms)

	lis, err := net.Listen("tcp", testAddr)
	assert.NoError(t, err)
	assert.NotNil(t, lis)
	if lis == nil {
		panic("listener cannot be nil")
	}
	go s.Serve(lis)
	return ms.q
}

type wrapper struct {
	server *mockServer
	stream Link_TalkServer
}

func newWrapper(s *mockServer, conn Link_TalkServer) flow.Conn {
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
