package link

import (
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"

	g "github.com/baetyl/baetyl-go/utils/protocol/grpc"
	"gopkg.in/tomb.v2"
)

// (中文注释后面提交会去掉，mock server的测试代码待补充)
// 整体行为：
// 1）在初始化后不会立即建立连接，会在第一次进行消息收发操作时打开连接
// 2）在调用异步消息收取方法 Receive 或同步消息发送方法 SendSync 后会开始监听收到的消息
//
// LinkerAPI Contact API
type LinkerAPI interface {
	// async mode
	Send(Source string, Destination string, content []byte) error
	Receive(handler Handler)
}

// Handler
// if []byte not nil or source is a sync message,
// will send a response to the source
type Handler func([]byte) []byte

type Linker struct {
	conn     *grpc.ClientConn
	cli      *Client
	stream   Link_TalkClient
	handler  Handler
	msgAsync chan *Message
	t        tomb.Tomb
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
		conn:     conn,
		cli:      cli,
		stream:   stream,
		handler:  handler,
		msgAsync: make(chan *Message),
	}
	l.receive()
	return l, nil
}

// receive 行为
// 1）执行后启动两个协程。协程1 进行消息的收取，解析和处理；协程2 消息响应数据的发送
// 2）协程1 会在收到消息后
//     a. 监听并接收消息
//     b. 判断是否handler是否可用
//        是->c
//        否->丢弃消息
//     c. 调用handler处理消息
//     d. 判断handler返回值是否为空
//        是->a
//        否->准备待响应消息
//     e. 向待发送消息chan写入响应信息
//
// receive implement Talk for receive async message
func (l *Linker) receive() {
	l.t.Go(func() error {
		for {
			in, err := l.stream.Recv()
			if err == io.EOF {
				close(l.msgAsync)
				return err
			}
			if err != nil {
				return err
			}
			fmt.Printf("receive stream = %v\n", in)
			if l.handler == nil {
				fmt.Println("handle not implemented, message dropped")
				continue
			}
			resp := l.handler(in.Content)
			// check : is sync message or handler return not nil
			msg := &Message{
				Content: resp,
				Context: &Context{
					ID:          uint64(time.Now().UnixNano()),
					TS:          uint64(time.Now().Unix()),
					QOS:         1,
					Flags:       0,
					Topic:       "$SYS/service/" + in.Context.Source,
					Source:      in.Context.Destination,
					Destination: in.Context.Source,
				},
			}
			select {
			case l.msgAsync <- msg:
			case <-l.t.Dying():
				close(l.msgAsync)
				return nil
			}
		}
	})
	l.t.Go(func() error {
		for {
			msg, ok := <-l.msgAsync
			if ok {
				err := l.stream.Send(msg)
				if err != nil {
					return err
				}
			}
		}
	})
}

// Send for send async message
func (l *Linker) Send(Source, Destination string, content []byte, timeout time.Duration) error {
	msg := &Message{
		Content: content,
		Context: &Context{
			ID:          uint64(time.Now().UnixNano()),
			TS:          uint64(time.Now().Unix()),
			QOS:         1,
			Flags:       0,
			Topic:       "$SYS/service/" + Destination,
			Source:      Source,
			Destination: Destination,
		},
	}
	return l.stream.Send(msg)
}

// Close closes Client
func (l *Linker) Close() error {
	if l.cli != nil {
		return l.conn.Close()
	}
	l.t.Kill(nil)
	return l.t.Wait()
}
