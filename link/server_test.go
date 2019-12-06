package link

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

type lkSerST struct{}

func (l *lkSerST) Call(ctx context.Context, msg *Message) (*Message, error) {
	return msg, nil
}

func (l *lkSerST) Talk(stream Link_TalkServer) error {
	for {
		in, err := stream.Recv()
		if err != nil {
			return err
		}
		if err = stream.Send(in); err != nil {
			return err
		}
	}
}

func TestLinkServer(t *testing.T) {
	scfg := ServerConfig{}
	scfg.Username = "svr"
	scfg.Password = "svr"
	scfg.Concurrent.Max = 10000
	scfg.MaxMessageSize = 100000
	scfg.CA = "./testcert/ca.pem"
	scfg.Cert = "./testcert/server.pem"
	scfg.Key = "./testcert/server.key"

	svr, err := NewServer(scfg)
	assert.NoError(t, err)
	defer svr.GracefulStop()
	s := &lkSerST{}
	RegisterLinkServer(svr, s)
	lis, err := net.Listen("tcp", "localhost:8080")
	assert.NoError(t, err)
	go svr.Serve(lis)
}
