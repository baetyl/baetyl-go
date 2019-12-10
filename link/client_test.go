package link

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/utils"

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
	certErrMsg = "rpc error: code = Unavailable desc =" +
		" all SubConns are in TransientFailure," +
		" latest connection error: connection error:" +
		" desc = \"transport: authentication handshake failed:" +
		" x509: certificate is valid for bd, not error\""

	accountErrMsg = "rpc error: code = Unauthenticated " +
		"desc = username or password not match"

	msgCall = &Message{
		Context: Context{
			ID:          0,
			TS:          1,
			QOS:         2,
			Flags:       3,
			Topic:       "$sys/service/cli",
			Source:      "cli",
			Destination: "ser",
		},
		Content: []byte("test msg call"),
	}

	msgCallResp = &Message{
		Context: Context{
			ID:          10,
			TS:          11,
			QOS:         12,
			Flags:       13,
			Topic:       "$sys/service/svr",
			Source:      "ser",
			Destination: "cli",
		},
		Content: []byte("test msg call resp"),
	}

	msgTalk = &Message{
		Context: Context{
			ID:          20,
			TS:          21,
			QOS:         22,
			Flags:       23,
			Topic:       "$sys/service/cli",
			Source:      "cli",
			Destination: "ser",
		},
		Content: []byte("test msg talk"),
	}

	msgTalkResp = &Message{
		Context: Context{
			ID:          30,
			TS:          31,
			QOS:         32,
			Flags:       33,
			Topic:       "$sys/service/svr",
			Source:      "svr",
			Destination: "cli",
		},
		Content: []byte("test msg talk resp"),
	}

	addr = "0.0.0.0:8273"

	scfg = ServerConfig{
		Auth: Auth{
			Account: Account{
				Username: "svr",
				Password: "svr",
			},
			Certificate: utils.Certificate{
				Cert: "./testcert/server.pem",
				Key:  "./testcert/server.key",
				CA:   "./testcert/ca.pem",
			},
		},
		MaxMessageSize: 4194304,
	}

	linkClientTests = []struct {
		name   string
		ccfg   ClientConfig
		params msg
		want   msg
		err    []errInfo
	}{
		{
			name: "Test 0 : Happy path",
			ccfg: ClientConfig{
				Address: "0.0.0.0:8273",
				Timeout: time.Duration(5) * time.Second,
				Auth: Auth{
					Account: Account{
						Username: "svr",
						Password: "svr",
					},
					Certificate: utils.Certificate{
						Cert: "./testcert/client.pem",
						Key:  "./testcert/client.key",
						CA:   "./testcert/ca.pem",
						Name: "bd",
					},
				},
				MaxMessageSize: 4194304,
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
			ccfg: ClientConfig{
				Address: "0.0.0.0:8273",
				Timeout: time.Duration(5) * time.Second,
				Auth: Auth{
					Account: Account{
						Username: "svr",
						Password: "svr",
					},
					Certificate: utils.Certificate{
						Cert: "./testcert/client.pem",
						Key:  "./testcert/server.key",
						CA:   "./testcert/ca.pem",
						Name: "bd",
					},
				},
				MaxMessageSize: 4194304,
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
			ccfg: ClientConfig{
				Address: "0.0.0.0:8273",
				Timeout: time.Duration(5) * time.Second,
				Auth: Auth{
					Account: Account{
						Username: "svr",
						Password: "error",
					},
					Certificate: utils.Certificate{
						Cert: "./testcert/client.pem",
						Key:  "./testcert/client.key",
						CA:   "./testcert/ca.pem",
						Name: "bd",
					},
				},
				MaxMessageSize: 4194304,
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

type lkSerCT struct {
	t *testing.T
}

func (l *lkSerCT) Call(ctx context.Context, msg *Message) (*Message, error) {
	checkMsg(l.t, msg, msgCall)
	return msgCallResp, nil
}

func (l *lkSerCT) Talk(stream Link_TalkServer) error {
	for {
		in, err := stream.Recv()
		if err != nil {
			return err
		}
		checkMsg(l.t, in, msgTalk)
		if err = stream.Send(msgTalkResp); err != nil {
			return err
		}
	}
}

func TestLinkClient(t *testing.T) {
	scfg.Concurrent.Max = 4194304
	ser, err := NewServer(scfg)
	assert.NoError(t, err)
	s := &lkSerCT{t: t}
	RegisterLinkServer(ser, s)
	lis, err := net.Listen("tcp", addr)
	assert.NoError(t, err)
	go ser.Serve(lis)

	wg := sync.WaitGroup{}
	x := 0
	for _, tt := range linkClientTests {
		t.Run(tt.name, func(t *testing.T) {
			cli, err := NewClient(tt.ccfg, nil)
			assert.Equal(t, tt.err[0].wantErr, err != nil)
			if cli != nil {
				resp, err := cli.Call(context.Background(), tt.params.msgCall)
				assert.Equal(t, tt.err[1].wantErr, err != nil)
				if err != nil {
					assert.Equal(t, accountErrMsg, err.Error())
				} else {
					checkMsg(t, msgCallResp, resp)
					stream, err := cli.Talk(context.Background())
					assert.NoError(t, err)
					go func() {
						in, err := stream.Recv()
						assert.NoError(t, err)
						checkMsg(t, in, msgTalkResp)
						x--
						wg.Done()
					}()
					err = stream.Send(tt.params.msgTalk)
					assert.Equal(t, tt.err[2].wantErr, err != nil)
					if err != nil {
						assert.Equal(t, accountErrMsg, err.Error())
					} else {
						wg.Add(1)
						x++
						err = stream.CloseSend()
						assert.NoError(t, err)
					}
				}
			}
		})
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
	assert.Equal(t, req.Context.Source, resp.Context.Source)
	assert.Equal(t, req.Context.Destination, resp.Context.Destination)
}
