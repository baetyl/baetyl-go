package context

import (
	"context"
	"fmt"
	"github.com/baetyl/baetyl-go/utils"
	"os"

	"github.com/baetyl/baetyl-go/api"
	"google.golang.org/grpc"
)

// Client client of api server
type Client struct {
	*api.Client
}

// NewEnvClient creates a new client by env
func NewEnvClient() (*Client, error) {
	addr := os.Getenv(EnvKeyAPIAddress)
	name := os.Getenv(EnvKeyServiceName)
	token := os.Getenv(EnvKeyServiceToken)
	if len(addr) == 0 {
		return nil, fmt.Errorf("Env (%s) not found", EnvKeyAPIAddress)
	}
	cc := api.ClientConfig{
		Address:  addr,
		Username: name,
		Password: token,
	}
	utils.SetDefaults(&cc)
	api, err := api.NewClient(cc)
	if err != nil {
		return nil, err
	}
	return &Client{
		Client: api,
	}, nil
}

// SetKV set kv
func (c *Client) SetKV(kv api.KV) error {
	_, err := c.KV.Set(context.Background(), &kv, grpc.WaitForReady(true))
	return err
}

// SetKVConext set kv which supports context
func (c *Client) SetKVConext(ctx context.Context, kv api.KV) error {
	_, err := c.KV.Set(ctx, &kv, grpc.WaitForReady(true))
	return err
}

// GetKV get kv
func (c *Client) GetKV(k []byte) (*api.KV, error) {
	return c.KV.Get(context.Background(), &api.KV{Key: k}, grpc.WaitForReady(true))
}

// GetKVConext get kv which supports context
func (c *Client) GetKVConext(ctx context.Context, k []byte) (*api.KV, error) {
	return c.KV.Get(ctx, &api.KV{Key: k}, grpc.WaitForReady(true))
}

// DelKV del kv
func (c *Client) DelKV(k []byte) error {
	_, err := c.KV.Del(context.Background(), &api.KV{Key: k}, grpc.WaitForReady(true))
	return err
}

// DelKVConext del kv which supports context
func (c *Client) DelKVConext(ctx context.Context, k []byte) error {
	_, err := c.KV.Del(ctx, &api.KV{Key: k}, grpc.WaitForReady(true))
	return err
}

// ListKV list kv with prefix
func (c *Client) ListKV(p []byte) ([]*api.KV, error) {
	kvs, err := c.KV.List(context.Background(), &api.KV{Key: p}, grpc.WaitForReady(true))
	if err != nil {
		return nil, err
	}
	return kvs.Kvs, nil
}

// ListKVContext list kv with prefix which supports context
func (c *Client) ListKVContext(ctx context.Context, p []byte) ([]*api.KV, error) {
	kvs, err := c.KV.List(ctx, &api.KV{Key: p}, grpc.WaitForReady(true))
	if err != nil {
		return nil, err
	}
	return kvs.Kvs, nil
}
