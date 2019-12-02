package link

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkServer(t *testing.T) {
	scfg := ServerConfig{}
	scfg.Address = "localhost:8080"
	scfg.Username = "svr"
	scfg.Password = "svr"
	scfg.Concurrent.Max = 10000
	scfg.CA = "./testcert/ca.pem"
	scfg.Cert = "./testcert/server.pem"
	scfg.Key = "./testcert/server.key"

	ser, err := NewServer(scfg, func(ctx context.Context, msg *Message) (message *Message, e error) {
		return msg, nil
	}, func(stream Link_TalkServer) error {
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
	assert.NoError(t, err)
	defer ser.Close()
}
