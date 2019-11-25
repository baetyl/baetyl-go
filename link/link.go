package link

import (
	"errors"
	"time"

	"github.com/baetyl/baetyl-go/utils/log"

	"google.golang.org/grpc"

	g "github.com/baetyl/baetyl-go/utils/protocol/grpc"
	"gopkg.in/tomb.v2"
)

const (
	FlagSync = 0x2
	FlagAck  = 0x4
)

// ErrClientClosed the client is closed
var ErrClientClosed = errors.New("grpc client is closed")

// (中文注释后面提交会去掉，mock server的测试代码待补充)
// 整体行为：
// 1）在初始化后不会立即建立连接，会在第一次进行消息收发操作时打开连接
// 2）在调用异步消息收取方法 Receive 或同步消息发送方法 SendSync 后会开始监听收到的消息

// Handler handler for reveice message
type Handler func([]byte) []byte

type Linker struct {
	conn      *grpc.ClientConn
	cli       *Client
	stream    Link_TalkClient
	handler   Handler
	msgAsync  chan *Message
	publisher *publisher
	t         tomb.Tomb
	log       *log.Logger
}

// NewLinker create client and start receive message
func NewLinker(cfg ClientConfig, handler Handler) (*Linker, error) {
	option := &g.ClientOption{}
	opts := option.Create().
		CredsFromFile(cfg.Certificate.Cert, cfg.Certificate.Name).
		CustomCred(&g.CustomCred{
			Username: cfg.Account.Username,
			Password: cfg.Account.Password,
		}).Build()

	conn, err := g.NewClientConnect(cfg.Address, cfg.Timeout, opts)
	if err != nil {
		return nil, err
	}
	cli := NewClient(conn)
	stream, err := cli.Talk()
	if err != nil {
		return nil, err
	}
	l := &Linker{
		conn:      conn,
		cli:       cli,
		stream:    stream,
		handler:   handler,
		msgAsync:  make(chan *Message),
		publisher: newPublisher(cfg.Timeout),
	}
	l.receive()
	return l, nil
}

// Close closes Client
func (l *Linker) Close() error {
	if l.cli != nil {
		return l.conn.Close()
	}
	return nil
}

func packetMsg(src, dest string, content []byte) *Message {
	return &Message{
		Content: content,
		Context: &Context{
			ID:          uint64(time.Now().UnixNano()),
			TS:          uint64(time.Now().Unix()),
			QOS:         1,
			Flags:       0,
			Topic:       "$SYS/service/" + dest,
			Source:      src,
			Destination: dest,
		},
	}
}

func packetAckMsg(in *Message, content []byte) *Message {
	return &Message{
		Content: content,
		Context: &Context{
			ID:          in.Context.ID,
			TS:          uint64(time.Now().Unix()),
			QOS:         1,
			Flags:       FlagAck,
			Topic:       "$SYS/service/" + in.Context.Source,
			Source:      in.Context.Destination,
			Destination: in.Context.Source,
		},
	}
}
