package link

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Client client of contact server
type Client struct {
	cli LinkClient
}

// NewClient creates a new client of functions server
func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		cli: NewLinkClient(conn),
	}
}

// Call sends request to link server
func (c *Client) Call(req *Message, timeout time.Duration) (*Message, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()
	return c.cli.Call(ctx, req, grpc.WaitForReady(false))
}

// Talk talk to link server
func (c *Client) Talk() (Link_TalkClient, error) {
	return c.cli.Talk(context.Background(), grpc.WaitForReady(false))
}
