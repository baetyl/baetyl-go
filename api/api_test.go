package api

import (
	"context"
	"github.com/baetyl/baetyl-go/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_APIServer(t *testing.T) {
	port := "51600"
	svr := FakeServer(t, port, nil)
	defer svr.GracefulStop()

	cliConf := ClientConfig{
		Address: "localhost:" + port,
	}
	utils.SetDefaults(&cliConf)
	cli, err := NewClient(cliConf)
	assert.NoError(t, err)
	assert.NotEmpty(t, cli)
	defer cli.Close()

	ctx := context.Background()
	_, err = cli.KV.Get(ctx, &KV{Key: []byte("aa")})
	assert.NoError(t, err)

	_, err = cli.KV.Set(ctx, &KV{Key: []byte("aa")})
	assert.NoError(t, err)

	_, err = cli.KV.Set(ctx, &KV{Key: []byte("aa"), Value: []byte("")})
	assert.NoError(t, err)

	_, err = cli.KV.Set(ctx, &KV{Key: []byte("aa"), Value: []byte("aadata")})
	assert.NoError(t, err)

	resp, err := cli.KV.Get(ctx, &KV{Key: []byte("aa")})
	assert.NoError(t, err)
	assert.Equal(t, resp.Key, []byte("aa"))
	assert.Equal(t, resp.Value, []byte("aadata"))

	_, err = cli.KV.Del(ctx, &KV{Key: []byte("aa")})
	assert.NoError(t, err)

	_, err = cli.KV.Del(ctx, &KV{Key: []byte("")})
	assert.NoError(t, err)

	resp, err = cli.KV.Get(ctx, &KV{Key: []byte("aa")})
	assert.NoError(t, err)
	assert.Equal(t, resp.Key, []byte("aa"))
	assert.Empty(t, resp.Value)

	_, err = cli.KV.Set(ctx, &KV{Key: []byte("/a"), Value: []byte("/root/a")})
	assert.NoError(t, err)

	_, err = cli.KV.Set(ctx, &KV{Key: []byte("/b"), Value: []byte("/root/b")})
	assert.NoError(t, err)

	_, err = cli.KV.Set(ctx, &KV{Key: []byte("/c"), Value: []byte("/root/c")})
	assert.NoError(t, err)

	respa, err := cli.KV.List(ctx, &KV{Key: []byte("/")})
	assert.NoError(t, err)
	assert.Len(t, respa.Kvs, 3)
}
