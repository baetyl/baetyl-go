package link

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkServer(t *testing.T) {
	scfg := LServerConfig{
		Address: "127.0.0.1:8080",
		Account: Account{Username: "svr", Password: "svr"},
		Certificate: LServerCert{
			Cert: "./testcert/server.pem",
			Key:  "./testcert/server.key",
		},
	}
	ser, err := NewLServer(scfg, func(ctx context.Context, msg *Message) (message *Message, e error) {
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
