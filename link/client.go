package link

import (
	"io"
	"time"

	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	QoS0 = 0
	QoS1 = 1
	QoS2 = 2

	FlagSync = 0x2
	FlagAck  = 0x4
)

type Handler func(*Message) error

// Client client of contact server
type Client struct {
	cfg  ClientConfig
	conn *grpc.ClientConn
	cli  LinkClient

	stream  Link_TalkClient
	handler Handler
	log     *log.Logger
	utils.Tomb
}

// NewClient creates a new client of functions server
func NewClient(cc ClientConfig, handler Handler) (*Client, error) {
	ctx, cel := context.WithTimeout(context.Background(), cc.Timeout)
	defer cel()

	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(int(cc.MaxMessageSize))),
	}
	tlsCfg, err := utils.NewTLSConfigClient(&cc.Certificate)
	if err != nil {
		return nil, err
	}
	if tlsCfg != nil {
		if !cc.InsecureSkipVerify {
			tlsCfg.ServerName = cc.Name
		}
		creds := credentials.NewTLS(tlsCfg)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}
	// Custom Credential
	opts = append(opts, grpc.WithPerRPCCredentials(&CustomCred{
		Data: map[string]string{
			KeyUsername: cc.Username,
			KeyPassword: cc.Password,
		},
	}))

	conn, err := grpc.DialContext(ctx, cc.Address, opts...)
	if err != nil {
		return nil, err
	}
	cli := &Client{
		cfg:     cc,
		conn:    conn,
		handler: handler,
		cli:     NewLinkClient(conn),
		log:     log.With(log.String("link", "client")),
	}
	stream, err := cli.Talk(context.Background())
	if err != nil {
		return nil, err
	}
	cli.stream = stream
	cli.Go(cli.receiving)
	return cli, nil
}

// Call sends request to link server
func (c *Client) Call(ctx context.Context, req *Message) (*Message, error) {
	return c.cli.Call(ctx, req, grpc.WaitForReady(true))
}

// Talk talk to link server
func (c *Client) Talk(ctx context.Context) (Link_TalkClient, error) {
	return c.cli.Talk(ctx, grpc.WaitForReady(true))
}

// Close closes the client
func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Send(src, dest string, qos uint32, content []byte) error {
	msg := packetMsg(src, dest, qos, content)
	return c.stream.Send(msg)
}

// receiving implement Talk for receive async message
func (c *Client) receiving() error {
	for {
		in, err := c.stream.Recv()
		if err == io.EOF {
			return err
		}
		if err != nil {
			c.log.Error("talk stream recv error", log.Error(err))
			return err
		}
		c.log.Debug("talk receive msg",
			log.String("src", in.Context.Source),
			log.String("dest", in.Context.Destination))

		// check : is ack message
		if (in.Context.Flags & FlagAck) == FlagAck {
			c.log.Debug("talk receive ack", log.Int("id", int(in.Context.ID)))
		} else {
			if c.handler != nil {
				err = c.handler(in)
				if err != nil {
					c.log.Error("handler exec error", log.Error(err))
				}
			} else {
				c.log.Warn("handler not implemented")
			}
			if !c.cfg.DisableAutoAck {
				msg := packetAckMsg(in)
				err = c.stream.Send(msg)
				if err != nil {
					return err
				}
			}
		}
	}
}

func packetMsg(src, dest string, qos uint32, content []byte) *Message {
	return &Message{
		Content: content,
		Context: &Context{
			ID:          uint64(time.Now().UnixNano()),
			TS:          uint64(time.Now().Unix()),
			QOS:         qos,
			Flags:       0,
			Topic:       "$SYS/service/" + dest,
			Source:      src,
			Destination: dest,
		},
	}
}

func packetAckMsg(in *Message) *Message {
	return &Message{
		Content: nil,
		Context: &Context{
			ID:          in.Context.ID,
			TS:          uint64(time.Now().Unix()),
			QOS:         QoS1,
			Flags:       FlagAck,
			Topic:       "$SYS/service/" + in.Context.Source,
			Source:      in.Context.Destination,
			Destination: in.Context.Source,
		},
	}
}
