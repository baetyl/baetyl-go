package link

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"gopkg.in/tomb.v2"
)

const (
	FlagSync = 0x2
	FlagResp = 0x4

	MsgTimeoutDefault = 30 * time.Second
)

// (中文注释后面提交会去掉，mock server的测试代码待补充)
// 整体行为：
// 1）在初始化后不会立即建立连接，会在第一次进行消息收发操作时打开连接
// 2）在调用异步消息收取方法 Receive 或同步消息发送方法 SendSync 后会开始监听收到的消息
//
// LinkerAPI Contact API
type LinkerAPI interface {
	// async mode
	Send(src string, dest string, content []byte) error
	Receive(handler Handler)
	// sync mode
	SendSync(src string, dest string,
		content []byte, timeout time.Duration) (*Message, error)
}

// Handler
// if []byte not nil or source is a sync message,
// will send a response to the source
type Handler func(*Message) ([]byte, error)

type Linker struct {
	cfg        LClientConfig
	cli        *LClient
	stream     Link_TalkClient
	handler    Handler
	handlerSem sync.WaitGroup
	msgAsync   chan *Message
	msgSync    cmap.ConcurrentMap
	t          tomb.Tomb
	once       sync.Once
}

func NewLinker(cfg LClientConfig) *Linker {
	return &Linker{
		cfg:        cfg,
		msgAsync:   make(chan *Message),
		msgSync:    cmap.New(),
		handlerSem: sync.WaitGroup{},
	}
}

// load create LClient for send/receive message
func load(l *Linker) error {
	cli, err := NewLClient(l.cfg)
	if err != nil {
		return err
	}
	l.cli = cli
	stream, err := cli.Talk()
	if err != nil {
		return err
	}
	l.stream = stream
	l.handlerSem.Add(1)
	return nil
}

// receive 行为
// 1）只会实际执行一次
// 2）执行后启动两个协程。协程1 进行消息的收取，解析和处理；协程2 进行异步或同步消息的响应数据的发送
// 3）协程1 会在收到消息后
//     a. 监听并接收消息
//     b. 判断是否是同步消息的响应消息：
//        是->查看消息是否过期，过期则丢弃；未过期则传回同步消息等待的chan中。转a
//        否->c
//     c. 判断是否handler是否可用
//        是->d
//        否->阻塞等待用户通过Receive传入可用handler
//     d. 调用handler处理消息
//     e. 判断消息是否 为同步消息或handler返回值不为空
//        是->准备待响应消息，转f
//        否->a
//     f. 判断是否是待响应的同步消息
//        是->ID设为和传入消息一致，FLAGS位设为FlagResp
//        否->生成ID
//     g. 向待发送消息chan写入响应信息
//
// receive implement Talk for receive async message
func (l *Linker) receive() {
	l.once.Do(func() {
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
				// check : is sync message relay
				if (in.Context.Flags & FlagResp) == FlagResp {
					msgKey := string(in.Context.ID)
					if l.msgSync.Has(msgKey) {
						item, ok := l.msgSync.Get(msgKey)
						if ok {
							item.(chan *Message) <- in
						} else {
							return errors.New("cmap msgSync get error")
						}
					} else {
						fmt.Printf("msg [id=%s] is dropped because of timeout", msgKey)
					}
					continue
				}
				if l.handler == nil {
					l.handlerSem.Wait()
				}
				resp, err := l.handler(in)
				if err != nil {
					fmt.Printf("Handler error = %s\n", err.Error())
					continue
				}
				// check : is sync message or handler return not nil
				if (in.Context.Flags&FlagSync) == FlagSync || resp != nil {
					msg := &Message{
						Content: resp,
						Context: &Context{
							TS:    uint64(time.Now().Unix()),
							QOS:   1,
							Flags: 0,
							Topic: "$sys/service/" + in.Context.Src,
							Src:   in.Context.Dest,
							Dest:  in.Context.Src,
						},
					}
					if (in.Context.Flags & FlagSync) == FlagSync {
						msg.Context.ID = in.Context.ID
						msg.Context.Flags = FlagResp
					} else {
						msg.Context.ID = uint64(time.Now().UnixNano())
					}
					select {
					case l.msgAsync <- msg:
					case <-l.t.Dying():
						close(l.msgAsync)
						return nil
					}
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
	})
}

// Receive set handler for processing messages
func (l *Linker) Receive(handler Handler) {
	l.receive()
	if l.handler == nil && handler != nil {
		l.handlerSem.Done()
	}
	if l.handler != nil && handler == nil {
		l.handlerSem.Add(1)
	}
	l.handler = handler
}

// Send for send async message
func (l *Linker) Send(src string, dest string, content []byte) error {
	if l.cli == nil {
		if err := load(l); err != nil {
			return err
		}
	}
	msg := &Message{
		Content: content,
		Context: &Context{
			ID:    uint64(time.Now().UnixNano()),
			TS:    uint64(time.Now().Unix()),
			QOS:   1,
			Flags: 0,
			Topic: "$sys/service/" + dest,
			Src:   src,
			Dest:  dest,
		},
	}
	return l.stream.Send(msg)
}

// SendSync send a sync message
func (l *Linker) SendSync(src string, dest string,
	content []byte, timeout time.Duration) (*Message, error) {
	if l.cli == nil {
		if err := load(l); err != nil {
			return nil, err
		}
	}
	msg := &Message{
		Content: content,
		Context: &Context{
			ID:    uint64(time.Now().UnixNano()),
			TS:    uint64(time.Now().Unix()),
			QOS:   1,
			Flags: 2,
			Topic: "$sys/service/" + dest,
			Src:   src,
			Dest:  dest,
		},
	}
	err := l.stream.Send(msg)
	if err != nil {
		return nil, err
	}
	relay := make(chan *Message)
	msgKey := string(msg.Context.ID)
	l.msgSync.Set(msgKey, relay)

	l.receive()

	deadline := time.Now().Add(timeout)
	if timeout <= 0 {
		timeout = MsgTimeoutDefault
	}
	var resp *Message
	timer := time.NewTimer(time.Until(deadline))
	select {
	case resp = <-relay:
	case <-timer.C:
		return nil, errors.New("timeout")
	}
	timer.Stop()
	return resp, nil
}

// Close closes LClient
func (l *Linker) Close() error {
	if l.cli != nil {
		return l.cli.Close()
	}
	l.t.Kill(nil)
	return l.t.Wait()
}
