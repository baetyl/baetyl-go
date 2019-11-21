package link

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type msg struct {
	msgCall *Message
	msgTalk *Message
}

type errInfo struct {
	wantErr bool
	errMsg  string
}

var (
	scfg = LServerConfig{
		Address: "0.0.0.0:8276",
		Account: Account{Username: "svr", Password: "svr"},
		Certificate: LServerCert{
			Cert: "./testcert/server.pem",
			Key:  "./testcert/server.key",
		},
	}

	accountErrMsg = "rpc error: code = Unauthenticated " +
		"desc = username or password not match"

	msgCall = &Message{
		Context: &Context{
			ID:    0,
			TS:    1,
			QOS:   2,
			Flags: 3,
			Topic: "$sys/service/cli",
			Src:   "cli",
			Dest:  "ser",
		},
		Content: []byte("test msg call"),
	}

	msgCallResp = &Message{
		Context: &Context{
			ID:    10,
			TS:    11,
			QOS:   12,
			Flags: 13,
			Topic: "$sys/service/svr",
			Src:   "ser",
			Dest:  "cli",
		},
		Content: []byte("test msg call resp"),
	}

	msgTalk = &Message{
		Context: &Context{
			ID:    20,
			TS:    21,
			QOS:   22,
			Flags: 23,
			Topic: "$sys/service/cli",
			Src:   "cli",
			Dest:  "ser",
		},
		Content: []byte("test msg talk"),
	}

	msgTalkResp = &Message{
		Context: &Context{
			ID:    30,
			TS:    31,
			QOS:   32,
			Flags: 33,
			Topic: "$sys/service/svr",
			Src:   "svr",
			Dest:  "cli",
		},
		Content: []byte("test msg talk resp"),
	}

	linkClientTests = []struct {
		name   string
		ccfg   LClientConfig
		params msg
		want   msg
		err    []errInfo
	}{
		{
			name: "Test 0 : Happy path",
			ccfg: LClientConfig{
				Address: "0.0.0.0:8276",
				Timeout: time.Duration(20) * time.Second,
				Account: Account{
					Username: "svr",
					Password: "svr",
				},
				Certificate: LClientCert{
					Cert: "./testcert/server.pem",
					Name: "bd",
				},
			},
			params: msg{
				msgCall: msgCall,
				msgTalk: msgTalk,
			},
			want: msg{
				msgCall: msgCallResp,
				msgTalk: msgTalkResp,
			},
			err: []errInfo{
				{wantErr: false},
				{wantErr: false},
				{wantErr: false},
			},
		},
		{
			name: "Test 1 : Cert error",
			ccfg: LClientConfig{
				Address: "0.0.0.0:8276",
				Timeout: time.Duration(1) * time.Second,
				Account: Account{
					Username: "svr",
					Password: "svr",
				},
				Certificate: LClientCert{
					Cert: "./testcert/server.pem",
					Name: "",
				},
			},
			params: msg{
				msgCall: msgCall,
				msgTalk: msgTalk,
			},
			want: msg{
				msgCall: msgCall,
				msgTalk: msgTalk,
			},
			err: []errInfo{
				{wantErr: true},
			},
		},
		{
			name: "Test 2 : Account error",
			ccfg: LClientConfig{
				Address: "0.0.0.0:8276",
				Timeout: time.Duration(1) * time.Second,
				Account: Account{
					Username: "svr",
					Password: "error",
				},
				Certificate: LClientCert{
					Cert: "./testcert/server.pem",
					Name: "bd",
				},
			},
			params: msg{
				msgCall: msgCall,
				msgTalk: msgTalk,
			},
			want: msg{
				msgCall: msgCallResp,
				msgTalk: msgTalkResp,
			},
			err: []errInfo{
				{wantErr: false},
				{
					wantErr: true,
					errMsg:  accountErrMsg,
				},
				{
					wantErr: true,
					errMsg:  accountErrMsg,
				},
			},
		},
	}
)

func TestLinkClient(t *testing.T) {
	ser, err := NewLServer(scfg, func(ctx context.Context, msg *Message) (message *Message, e error) {
		checkMsg(t, msg, msgCall)
		return msgCallResp, nil
	}, func(stream Link_TalkServer) error {
		for {
			in, err := stream.Recv()
			if err != nil {
				return err
			}
			checkMsg(t, in, msgTalk)
			if err = stream.Send(msgTalkResp); err != nil {
				return err
			}
		}
	})
	assert.NoError(t, err)
	defer ser.Close()

	wg := sync.WaitGroup{}
	for _, tt := range linkClientTests {
		cli, err := NewLClient(tt.ccfg)
		assert.Equal(t, tt.err[0].wantErr, err != nil)
		if cli != nil {
			resp, err := cli.Call(tt.params.msgCall)
			assert.Equal(t, tt.err[1].wantErr, err != nil)
			if err != nil {
				assert.Equal(t, accountErrMsg, err.Error())
				continue
			}
			checkMsg(t, msgCallResp, resp)
			stream, err := cli.Talk()
			assert.NoError(t, err)
			go func() {
				in, err := stream.Recv()
				assert.NoError(t, err)
				checkMsg(t, in, msgTalkResp)
				wg.Done()
			}()
			err = stream.Send(tt.params.msgTalk)
			assert.Equal(t, tt.err[2].wantErr, err != nil)
			if err != nil {
				assert.Equal(t, accountErrMsg, err.Error())
				continue
			}
			wg.Add(1)
			err = stream.CloseSend()
			assert.NoError(t, err)
		}
	}
	wg.Wait()
}

func checkMsg(t *testing.T, req *Message, resp *Message) {
	assert.Equal(t, string(req.Content), string(resp.Content))
	assert.Equal(t, req.Context.ID, resp.Context.ID)
	assert.Equal(t, req.Context.TS, resp.Context.TS)
	assert.Equal(t, req.Context.QOS, resp.Context.QOS)
	assert.Equal(t, req.Context.Flags, resp.Context.Flags)
	assert.Equal(t, req.Context.Topic, resp.Context.Topic)
	assert.Equal(t, req.Context.Src, resp.Context.Src)
	assert.Equal(t, req.Context.Dest, resp.Context.Dest)
}
