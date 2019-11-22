package link

import (
	"fmt"

	g "github.com/baetyl/baetyl-go/utils/protocol/grpc"
	"golang.org/x/net/context"
)

// Call message handler
type Call func(context.Context, *Message) (*Message, error)

// Talk stream message handler
type Talk func(Link_TalkServer) error

// Server Link server to handle message
type Server struct {
	call     Call
	talk     Talk
	username string
	password string
}

// NewServer creates a new Link server
func NewServer(username, password string, call Call, talk Talk) *Server {
	s := &Server{
		call:     call,
		talk:     talk,
		username: username,
		password: password,
	}
	return s
}

// Call handles message
func (s *Server) Call(c context.Context, m *Message) (*Message, error) {
	if s.call == nil {
		return nil, fmt.Errorf("call handle not implemented")
	}
	if authResult, err := g.Authenticate(c, s.username, s.password); !authResult {
		return nil, err
	}
	return s.call(c, m)
}

// Talk stream message handler
func (s *Server) Talk(stream Link_TalkServer) error {
	if s.talk == nil {
		return fmt.Errorf("talk handle not implemented")
	}
	if authResult, err := g.Authenticate(stream.Context(), s.username, s.password); !authResult {
		return err
	}
	return s.talk(stream)
}
