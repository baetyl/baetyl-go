package baetyl

import (
	"github.com/baetyl/baetyl-go/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var callOpt = grpc.WaitForReady(false)

// CClient client of contact server
type CClient struct {
	cfg  ContactClientConfig
	conn *grpc.ClientConn
	cli  ContactClient
}

// GetRequestMetadata & RequireTransportSecurity for Custom Credential
// GetRequestMetadata gets the current request metadata, refreshing tokens if required
func (c *CClient) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"username": c.cfg.Auth.Username,
		"password": c.cfg.Auth.Password,
	}, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security
func (c *CClient) RequireTransportSecurity() bool {
	return len(c.cfg.Auth.Username) > 0
}

// NewCClient creates a new client of functions server
func NewCClient(cc ContactClientConfig) (*CClient, error) {
	ctx, cel := context.WithTimeout(context.Background(), cc.Timeout)
	defer cel()
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithBackoffMaxDelay(cc.Backoff.Max),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(int(cc.Message.Length.Max))),
	}

	tls, err := utils.NewTLSServerConfig(cc.Auth.Certificate)
	if err != nil {
		return nil, err
	}
	if tls != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tls)))
	}

	conn, err := grpc.DialContext(ctx, cc.Address, opts...)
	if err != nil {
		return nil, err
	}
	return &CClient{
		cfg:  cc,
		conn: conn,
		cli:  NewContactClient(conn),
	}, nil
}

// Call sends request to contact server
func (c *CClient) Call(req *Message) (*Message, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), c.cfg.Timeout)
	defer cancel()
	return c.cli.Callback(ctx, req, callOpt)
}

// Talk talk to contact server
func (c *CClient) Talk() (Contact_TalkClient, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), c.cfg.Timeout)
	defer cancel()
	return c.cli.Talk(ctx, callOpt)
}

// Close closes the client
func (c *CClient) Close() error {
	return c.conn.Close()
}
