package context

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/api"
	"github.com/baetyl/baetyl-go/link"
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestNewEnvClient(t *testing.T) {
	cli, err := NewEnvClient()
	assert.EqualError(t, err, "Env (BAETYL_API_ADDRESS) not found")
	assert.Nil(t, cli)

	port := "52006"
	svr := api.FakeServer(t, port, new(mockAuthenticator))

	// new
	os.Setenv(EnvKeyAPIAddress, "localhost:"+port)
	os.Setenv(EnvKeyServiceName, "baetyl")
	os.Setenv(EnvKeyServiceToken, "baetyl")
	cli, err = NewEnvClient()
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	a := api.KV{
		Key:   []byte("name"),
		Value: []byte("baetyl"),
	}
	_, err = cli.GetKV(a.Key)
	assert.NoError(t, err)

	err = cli.SetKV(a)
	assert.NoError(t, err)

	resp, err := cli.GetKV(a.Key)
	assert.NoError(t, err)
	assert.Equal(t, resp.Value, a.Value)

	err = cli.DelKV(a.Key)
	assert.NoError(t, err)

	err = cli.SetKV(a)
	assert.NoError(t, err)

	a.Key = []byte("bb")
	err = cli.SetKV(a)
	assert.NoError(t, err)

	respa, err := cli.ListKV([]byte(""))
	assert.NoError(t, err)
	assert.Len(t, respa, 2)

	ctx0, cel0 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cel0()
	err = cli.SetKVConext(ctx0, a)
	assert.NoError(t, err)

	svr.GracefulStop()

	ctx1, cel1 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cel1()
	_, err = cli.GetKVConext(ctx1, a.Key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DeadlineExceeded desc")

	ctx2, cel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cel2()
	err = cli.SetKVConext(ctx2, a)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DeadlineExceeded desc")

	ctx3, cel3 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cel3()
	resp, err = cli.GetKVConext(ctx3, a.Key)
	assert.Contains(t, err.Error(), "DeadlineExceeded desc")

	ctx4, cel4 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cel4()
	err = cli.DelKVConext(ctx4, a.Key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DeadlineExceeded desc")

	ctx5, cel5 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cel5()
	respa, err = cli.ListKVContext(ctx5, []byte(""))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DeadlineExceeded desc")

	svr = api.FakeServer(t, port, new(mockAuthenticator))

	a.Key = []byte("aa")
	err = cli.SetKV(a)
	assert.NoError(t, err)

	ctx6, cel6 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cel6()
	resp, err = cli.GetKVConext(ctx6, a.Key)
	assert.NoError(t, err)
	assert.Equal(t, resp.Value, a.Value)

	ctx7, cel7 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cel7()
	err = cli.DelKVConext(ctx7, a.Key)
	assert.NoError(t, err)

	ctx8, cel8 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cel8()
	err = cli.SetKVConext(ctx8, a)
	assert.NoError(t, err)

	ctx9, cel9 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cel9()
	a.Key = []byte("bb")
	err = cli.SetKVConext(ctx9, a)
	assert.NoError(t, err)

	ctx10, cel10 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cel10()
	respa, err = cli.ListKVContext(ctx10, []byte(""))
	assert.NoError(t, err)
	assert.Len(t, respa, 2)

	svr.GracefulStop()
	cli.Close()
}

type mockMaster struct{}

func (*mockMaster) Auth(u, p string) bool {
	if u == "baetyl" && p == "baetyl" {
		return true
	}
	return false
}

// KVService kv server
type mockKVService struct {
	m map[string][]byte
}

// Set set kv
func (s *mockKVService) Set(ctx context.Context, kv *api.KV) (*types.Empty, error) {
	s.m[string(kv.Key)] = kv.Value
	return new(types.Empty), nil
}

// Get get kv
func (s *mockKVService) Get(ctx context.Context, kv *api.KV) (*api.KV, error) {
	return &api.KV{
		Key:   kv.Key,
		Value: s.m[string(kv.Key)],
	}, nil
}

// Del del kv
func (s *mockKVService) Del(ctx context.Context, kv *api.KV) (*types.Empty, error) {
	delete(s.m, string(kv.Key))
	return new(types.Empty), nil
}

// List list kvs with prefix
func (s *mockKVService) List(ctx context.Context, kv *api.KV) (*api.KVs, error) {
	kvs := api.KVs{
		Kvs: []*api.KV{},
	}
	for k, v := range s.m {
		kvs.Kvs = append(kvs.Kvs, &api.KV{
			Key:   []byte(k),
			Value: v,
		})
	}
	return &kvs, nil
}

type mockAuthenticator struct{}

func (auth mockAuthenticator) Authenticate(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return link.ErrUnauthenticated
	}
	var u, p string
	if val, ok := md[link.KeyUsername]; ok {
		u = val[0]
	}
	if val, ok := md[link.KeyPassword]; ok {
		p = val[0]
	}
	if u != "baetyl" || p != "baetyl" {
		return link.ErrUnauthenticated
	}
	return nil
}
