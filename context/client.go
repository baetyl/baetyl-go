package context

import (
	"context"
	"fmt"
	"os"

	"github.com/baetyl/baetyl-go/kv"
	"github.com/baetyl/baetyl-go/link"
	"github.com/baetyl/baetyl-go/utils"
	"google.golang.org/grpc"
)

// NewEnvClient creates a new client by env
func NewEnvClient() (*Client, error) {
	addr := os.Getenv(EnvKeyAPIAddress)
	name := os.Getenv(EnvKeyServiceName)
	token := os.Getenv(EnvKeyServiceToken)
	if len(addr) == 0 {
		return nil, fmt.Errorf("Env (%s) not found", EnvKeyAPIAddress)
	}
	conf := link.ClientConfig{
		Address:  addr,
		Username: name,
		Password: token,
	}
	utils.SetDefaults(&conf)
	conn, err := link.NewClientConn(conf)
	if err != nil {
		return nil, err
	}
	kv := kv.NewKVServiceClient(conn)
	return &Client{
		conf: conf,
		conn: conn,
		KV:   kv,
	}, nil
}

// SetKV set kv
func (c *Client) SetKV(kv kv.KV) error {
	_, err := c.KV.Set(context.Background(), &kv, grpc.WaitForReady(true))
	return err
}

// SetKVConext set kv which supports context
func (c *Client) SetKVConext(ctx context.Context, kv kv.KV) error {
	_, err := c.KV.Set(ctx, &kv, grpc.WaitForReady(true))
	return err
}

// GetKV get kv
func (c *Client) GetKV(k []byte) (*kv.KV, error) {
	return c.KV.Get(context.Background(), &kv.KV{Key: k}, grpc.WaitForReady(true))
}

// GetKVConext get kv which supports context
func (c *Client) GetKVConext(ctx context.Context, k []byte) (*kv.KV, error) {
	return c.KV.Get(ctx, &kv.KV{Key: k}, grpc.WaitForReady(true))
}

// DelKV del kv
func (c *Client) DelKV(k []byte) error {
	_, err := c.KV.Del(context.Background(), &kv.KV{Key: k}, grpc.WaitForReady(true))
	return err
}

// DelKVConext del kv which supports context
func (c *Client) DelKVConext(ctx context.Context, k []byte) error {
	_, err := c.KV.Del(ctx, &kv.KV{Key: k}, grpc.WaitForReady(true))
	return err
}

// ListKV list kv with prefix
func (c *Client) ListKV(p []byte) ([]*kv.KV, error) {
	kvs, err := c.KV.List(context.Background(), &kv.KV{Key: p}, grpc.WaitForReady(true))
	if err != nil {
		return nil, err
	}
	return kvs.Kvs, nil
}

// ListKVContext list kv with prefix which supports context
func (c *Client) ListKVContext(ctx context.Context, p []byte) ([]*kv.KV, error) {
	kvs, err := c.KV.List(ctx, &kv.KV{Key: p}, grpc.WaitForReady(true))
	if err != nil {
		return nil, err
	}
	return kvs.Kvs, nil
}

// Close closes the client
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
