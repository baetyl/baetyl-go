package websocket

import (
	"net/http"
	"net/url"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/gorilla/websocket"
	"github.com/panjf2000/ants/v2"
)

type Client struct {
	conn    chan *websocket.Conn
	u       url.URL
	dialer  websocket.Dialer
	antPool *ants.Pool
}

func NewClient(ops *ClientOptions, readMsgChan []chan *ReadMsg) (*Client, error) {
	u := url.URL{Scheme: ops.Schema, Host: ops.Address, Path: ops.Path}
	dialer := websocket.Dialer{
		NetDial:          nil,
		NetDialContext:   nil,
		Proxy:            http.ProxyFromEnvironment,
		TLSClientConfig:  ops.TLSConfig,
		HandshakeTimeout: ops.TLSHandshakeTimeout,
	}
	p, _ := ants.NewPool(1)
	if ops.SyncMaxConcurrency != 0 {
		p, _ = ants.NewPool(ops.SyncMaxConcurrency)
	}
	if readMsgChan != nil && cap(readMsgChan) < ops.SyncMaxConcurrency {
		return nil, errors.New("read msg cap must > SyncMaxConcurrency")
	}
	connect := make(chan *websocket.Conn, ops.SyncMaxConcurrency)

	// 根据设置创建连接池
	for i := 0; i < ops.SyncMaxConcurrency; i++ {
		con, _, err := dialer.Dial(u.String(), nil)
		if err != nil {
			return nil, err
		}
		// 每个链接创建一个协程readMsg
		if readMsgChan != nil {
			go ReadConMsg(con, readMsgChan[i])
		}
		connect <- con
	}

	return &Client{
		conn:    connect,
		u:       u,
		dialer:  dialer,
		antPool: p,
	}, nil
}

func (c *Client) SendMsg(msg []byte) error {
	con := <-c.conn
	err := con.WriteMessage(websocket.TextMessage, msg)
	c.conn <- con
	return err
}

func ReadConMsg(con *websocket.Conn, readMsg chan *ReadMsg) {
	for {
		msgType, data, err := con.ReadMessage()
		msg := &ReadMsg{
			MsgType: msgType,
			Data:    data,
			Err:     err,
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
			syncResult <- &SyncResults{
				Err:      err,
				SendCost: sendElapsed,
				SyncCost: syncElapsed,
				Extra:    extra,
			}
			if err != nil {

			}
		})
	if err != nil {
		syncResult <- &SyncResults{
			Err:      err,
			SendCost: 0,
			SyncCost: 0,
			Extra:    extra,
		}
	}
}
