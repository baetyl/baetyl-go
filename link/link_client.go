package link

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// LClient client of contact server
type LClient struct {
	cfg  LClientConfig
	conn *grpc.ClientConn
	cli  LinkClient
}

type customCred struct {
	ac Account
}

// GetRequestMetadata & RequireTransportSecurity for Custom Credential
// GetRequestMetadata gets the current request metadata, refreshing tokens if required
func (c *customCred) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"username": c.ac.Username,
		"password": c.ac.Password,
	}, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security
func (c *customCred) RequireTransportSecurity() bool {
	return len(c.ac.Username) > 0
}

// NewLClient creates a new client of functions server
func NewLClient(cc LClientConfig) (*LClient, error) {
	ctx, cel := context.WithTimeout(context.Background(), cc.Timeout)
	defer cel()

	opts := []grpc.DialOption{grpc.WithBlock()}

	if cc.Certificate.Insecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		creds, err := credentials.NewClientTLSFromFile(cc.Certificate.Cert, cc.Certificate.Name)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}
	// Custom Credential
	opts = append(opts, grpc.WithPerRPCCredentials(&customCred{ac: cc.Account}))

	conn, err := grpc.DialContext(ctx, cc.Address, opts...)
	if err != nil {
		return nil, err
	}
	return &LClient{
		cfg:  cc,
		conn: conn,
		cli:  NewLinkClient(conn),
	}, nil
}

// Call sends request to link server
func (c *LClient) Call(req *Message) (*Message, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), c.cfg.Timeout)
	defer cancel()
	return c.cli.Call(ctx, req, grpc.WaitForReady(false))
}

// Talk talk to link server
func (c *LClient) Talk() (Link_TalkClient, error) {
	return c.cli.Talk(context.Background(), grpc.WaitForReady(false))
}

// Close closes the client
func (c *LClient) Close() error {
	return c.conn.Close()
}
