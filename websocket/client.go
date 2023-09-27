package websocket

import (
	"net/http"
	"net/url"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/gorilla/websocket"
	"github.com/panjf2000/ants/v2"
)

type WsConnect struct {
	wscon       *websocket.Conn
	readMsgChan chan *v1.Message
}

type Client struct {
	pool        chan *WsConnect
	connNum     int
	u           url.URL
	dialer      websocket.Dialer
	antPool     *ants.Pool
	ops         *ClientOptions
	readMsgChan []chan *v1.Message
	log         *log.Logger
}

// NewClient 函数用于创建一个Client对象
// readMsgChan 为读取信息的通道 需要配置和并发数量一致
func NewClient(ops *ClientOptions, readMsgChan []chan *v1.Message) (*Client, error) {
	u := url.URL{Scheme: ops.Schema, Host: ops.Address, Path: ops.Path}
	dialer := websocket.Dialer{
		NetDial:          nil,
		NetDialContext:   nil,
		Proxy:            http.ProxyFromEnvironment,
		TLSClientConfig:  ops.TLSConfig,
		HandshakeTimeout: ops.TLSHandshakeTimeout,
	}
	// 最少为1条
	if ops.SyncMaxConcurrency <= 0 {
		ops.SyncMaxConcurrency = 1
	}

	p, err := ants.NewPool(ops.SyncMaxConcurrency)
	if err != nil {
		return nil, err
	}

	if readMsgChan != nil && cap(readMsgChan) < ops.SyncMaxConcurrency {
		return nil, errors.New("read msg cap must > SyncMaxConcurrency")
	}
	connect := make(chan *WsConnect, ops.SyncMaxConcurrency)
	client := &Client{
		pool:        connect,
		connNum:     0,
		u:           u,
		dialer:      dialer,
		antPool:     p,
		ops:         ops,
		readMsgChan: readMsgChan,
		log:         log.L().With(log.Any("link", "websocket link")),
	}
	go client.initLink()
	return client, nil
}

func (c *Client) initLink() {
	// 根据设置创建连接池
	for i := 0; i < c.ops.SyncMaxConcurrency; i++ {
		var connectReadMsgChan chan *v1.Message = nil
		if c.readMsgChan != nil {
			connectReadMsgChan = c.readMsgChan[i]
		}
		ws, err := c.Connect(connectReadMsgChan)
		if err != nil {
			c.log.Error("link websocket error", log.Any("err", err))
		}
		// 为了保证连接池数量 失败wscon 以nil方式放入连接池 每次发送的时候重新连接
		c.pool <- ws
	}
}

func (c *Client) Connect(readMsgChan chan *v1.Message) (*WsConnect, error) {
	con, _, err := c.dialer.Dial(c.u.String(), nil)
	ws := &WsConnect{
		readMsgChan: readMsgChan,
	}
	if err != nil {
		ws.wscon = nil
		c.log.Error("websocket link error", log.Any("err", err))
		return ws, err
	} else {
		ws.wscon = con
		if c.readMsgChan != nil {
			go ws.ReadConMsg(readMsgChan)
		}
	}
	return ws, nil
}

func (c *Client) SendMsg(msg []byte) error {
	con := <-c.pool
	var err error
	if con.wscon == nil {
		con, err = c.Connect(con.readMsgChan)
		if err != nil {
			c.pool <- con
			c.log.Error("retry link websocket  error", log.Any("err", err))
			return err
		}
	}
	err = con.wscon.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		c.log.Error("websocket write msg error", log.Any("err", err))
		con, err = c.Connect(con.readMsgChan)
		if err != nil {
			c.pool <- con
			c.log.Error("retry link websocket  error", log.Any("err", err))
			return err
		}
	}
	c.pool <- con
	return err
}

func (w *WsConnect) ReadConMsg(readMsg chan *v1.Message) {
	for {
		if w.wscon == nil {
			return
		}
		msgType, data, err := w.wscon.ReadMessage()
		msg := &v1.Message{}
		if err != nil {
			msg = &v1.Message{
				Kind:     v1.MessageError,
				Metadata: nil,
				Content:  v1.LazyValue{Value: err},
			}
		} else {
			msg = &v1.Message{
				Kind:     v1.MessageWebsocketRead,
				Metadata: nil,
				Content: v1.LazyValue{Value: WebsocketReadMsg{
					MsgType: msgType,
					Data:    data,
				}},
			}
		}

		select {
		case readMsg <- msg:
		default:
			log.Error(errors.New("can not add  msg to readMsg from websocket con"))
		}
	}
}

func (c *Client) SyncSendMsg(msg []byte, syncResult chan *SyncResults, extra map[string]interface{}) {
	SyncSendStart := time.Now()
	err := c.antPool.Submit(
		func() {
			sendStart := time.Now()
			err := c.SendMsg(msg)
			sendElapsed := time.Since(sendStart)
			syncElapsed := time.Since(SyncSendStart)
			result := &SyncResults{
				Err:      err,
				SendCost: sendElapsed,
				SyncCost: syncElapsed,
				Extra:    extra,
			}
			select {
			case syncResult <- result:
			default:
				log.Error(errors.New("can not add send result to syncResult from websocket con"))
			}
		})
	if err != nil {
		result := &SyncResults{
			Err:      err,
			SendCost: 0,
			SyncCost: 0,
			Extra:    extra,
		}
		select {
		case syncResult <- result:
		default:
			log.Error(errors.New("can not add send result to syncResult from websocket con"))
		}
	}
}
