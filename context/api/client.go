package api

import (
	"github.com/baetyl/baetyl-go/link"
	"github.com/baetyl/baetyl-go/utils"
)

// NewClient creates a new client
func NewClient(conf ClientConfig) (*Client, error) {
	utils.SetDefaults(&conf)
	conn, err := link.NewClientConn(link.ClientConfig(conf))
	if err != nil {
		return nil, err
	}
	kv := NewKVServiceClient(conn)
	return &Client{
		conf: conf,
		conn: conn,
		KV:   kv,
	}, nil
}

// Close closes the client
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
