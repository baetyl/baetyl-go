package link

import (
	"github.com/baetyl/baetyl-go/link/auth"
	"github.com/baetyl/baetyl-go/utils"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Client client of contact server
type Client struct {
	cfg  ClientConfig
	conn *grpc.ClientConn
	cli  LinkClient
}

// NewClient creates a new client of functions server
func NewClient(cc ClientConfig) (*Client, error) {
	ctx, cel := context.WithTimeout(context.Background(), cc.Timeout)
	defer cel()

	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				MaxDelay: cc.Backoff.Max,
			},
		}),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(int(cc.Message.Length.Max))),
	}
	tlsCfg, err := utils.NewTLSConfigClient(&cc.Certificate)
	if err != nil {
		return nil, err
	}
	if tlsCfg != nil {
		tlsCfg.ServerName = cc.Name
		creds := credentials.NewTLS(tlsCfg)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}
	// Custom Credential
	opts = append(opts, grpc.WithPerRPCCredentials(&auth.CustomCred{
		Data: map[string]string{
			auth.KeyUsername: cc.Username,
			auth.KeyPassword: cc.Password,
		},
	}))

	conn, err := grpc.DialContext(ctx, cc.Address, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		cfg:  cc,
		conn: conn,
		cli:  NewLinkClient(conn),
	}, nil
}

// Call sends request to link server
func (c *Client) Call(ctx context.Context, req *Message) (*Message, error) {
	return c.cli.Call(ctx, req, grpc.WaitForReady(false))
}

// Talk talk to link server
func (c *Client) Talk(ctx context.Context) (Link_TalkClient, error) {
	return c.cli.Talk(ctx, grpc.WaitForReady(false))
}

// Close closes the client
func (c *Client) Close() error {
	return c.conn.Close()
}
