package link

import (
	"fmt"
	"io"
	"time"

	"github.com/baetyl/baetyl-go/utils"
	"google.golang.org/grpc/credentials"
	"gopkg.in/tomb.v2"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
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

	stream   Link_TalkClient
	handler  Handler
	ack      chan *Message
	asyncMsg chan *Message
	t        tomb.Tomb
}

// NewClient creates a new client of functions server
func NewClient(cc ClientConfig, handler Handler) (*Client, error) {
	ctx, cel := context.WithTimeout(context.Background(), cc.Timeout)
	defer cel()

	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(int(cc.MaxSize))),
	}
	if cc.Insecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		tlsCfg, err := utils.NewTLSConfigClient(&cc.Certificate)
		if err != nil {
			return nil, err
		}
		if tlsCfg != nil {
			tlsCfg.InsecureSkipVerify = true // don't verifies the server's certificate chain and host name
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
	}

	conn, err := grpc.DialContext(ctx, cc.Address, opts...)
	if err != nil {
		return nil, err
	}
	cli := &Client{
		cfg:     cc,
		conn:    conn,
		handler: handler,
		cli:     NewLinkClient(conn),
	}
	stream, err := cli.Talk(context.Background())
	if err != nil {
		return nil, err
	}
	cli.stream = stream
	cli.t.Go(cli.publish)
	cli.t.Go(cli.receive)
	return cli, nil
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

func (c *Client) Send(src, dest string, qos uint32, content []byte) error {
	msg := packetMsg(src, dest, qos, content)
	err := c.stream.Send(msg)
	if err != nil {
		return err
	}
	if qos != QoS0 {
		a := <-c.ack
		if a.Context.ID != msg.Context.ID {
			return fmt.Errorf("Mseeage ack error."+
				" Expect id : %d, Actual id : %d\n", msg.Context.ID, a.Context.ID)
		}
	}
	return nil
}

// publish send message
func (c *Client) publish() error {
	for {
		msg, ok := <-c.asyncMsg
		if ok {
			err := c.stream.Send(msg)
			if err != nil {
				return err
			}
		}
	}
}

// receive implement Talk for receive async message
func (c *Client) receive() error {
	for {
		in, err := c.stream.Recv()
		if err == io.EOF {
			close(c.ack)
			return err
		}
		if err != nil {
			return err
		}
		// check : is ack message
		if (in.Context.Flags & FlagAck) == FlagAck {
			c.ack <- in
		} else {
			if c.handler != nil {
				err = c.handler(in)
				if err != nil {
					fmt.Printf("handler exec error, err = %v\n", err.Error())
				}
			} else {
				fmt.Println("handler not implemented, ack context is null")
			}
			if c.cfg.Ack {
				msg := packetAckMsg(in)
				c.asyncMsg <- msg
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
