package link

import (
	"context"
	"testing"

	"google.golang.org/grpc"

	g "github.com/baetyl/baetyl-go/utils/protocol/grpc"
	"github.com/stretchr/testify/assert"
)

func TestLinkServer(t *testing.T) {
	option := &g.ServerOption{}
	opts := option.
		Create().
		CredsFromFile("./testcert/server.pem", "./testcert/server.key").
		Build()
	s := NewServer("test",
		"123456",
		func(c context.Context, msg *Message) (*Message, error) {
			return msg, nil
		},
		func(stream Link_TalkServer) error {
			for {
				in, err := stream.Recv()
				if err != nil {
					return err
				}
				if err = stream.Send(in); err != nil {
					return err
				}
			}
		})
	svr, err := g.NewServer(g.NetTCP, "0.0.0.0:8273", opts, func(svr *grpc.Server) {
		RegisterLinkServer(svr, s)
	})
	assert.NoError(t, err)

	assert.NoError(t, err)
	defer svr.GracefulStop()
}
