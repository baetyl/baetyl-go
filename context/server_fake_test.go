package context

import (
	"context"
	"net"
	"testing"

	"github.com/baetyl/baetyl-go/api"
	"github.com/baetyl/baetyl-go/link"
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// FakeServer the fake of link server for test only
func FakeServer(t *testing.T, port string, auth link.Authenticator) *grpc.Server {
	var opts []grpc.ServerOption
	if auth != nil {
		ui := func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			err := auth.Authenticate(ctx)
			assert.NoError(t, err)
			return handler(ctx, req)
		}
		opts = append(opts, grpc.UnaryInterceptor(ui))
	}
	svr := grpc.NewServer(opts...)

	api.RegisterKVServiceServer(svr, &mockKVService{
		m: make(map[string][]byte),
	})
	lis, err := net.Listen("tcp", ":"+port)
	assert.NoError(t, err)
	go svr.Serve(lis)
	return svr
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
