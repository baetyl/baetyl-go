package context

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/kv"
	"github.com/stretchr/testify/assert"
)

func TestNewEnvClient(t *testing.T) {
	cli, err := NewEnvClient()
	assert.EqualError(t, err, "Env (BAETYL_API_ADDRESS) not found")
	assert.Nil(t, cli)

	port := "52006"
	svr := FakeServer(t, port, new(mockAuthenticator))

	// new
	os.Setenv(EnvKeyAPIAddress, "localhost:"+port)
	os.Setenv(EnvKeyServiceName, "baetyl")
	os.Setenv(EnvKeyServiceToken, "baetyl")
	cli, err = NewEnvClient()
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	a := kv.KV{
		Key:   "name",
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

	a.Key = "bb"
	err = cli.SetKV(a)
	assert.NoError(t, err)

	respa, err := cli.ListKV("")
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
	respa, err = cli.ListKVContext(ctx5, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DeadlineExceeded desc")

	svr = FakeServer(t, port, new(mockAuthenticator))

	a.Key = "aa"
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
	a.Key = "bb"
	err = cli.SetKVConext(ctx9, a)
	assert.NoError(t, err)

	ctx10, cel10 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cel10()
	respa, err = cli.ListKVContext(ctx10, "")
	assert.NoError(t, err)
	assert.Len(t, respa, 2)

	svr.GracefulStop()
	cli.Close()
}
