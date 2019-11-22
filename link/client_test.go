package link

import (
	"context"
	"sync"
	"testing"
	"time"

	g "github.com/baetyl/baetyl-go/utils/protocol/grpc"
	"google.golang.org/grpc"

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
		Context: &Context{
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
		Context: &Context{
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
		Context: &Context{
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
		Context: &Context{
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
			ccfg: ClientConfig{
				Address: "0.0.0.0:8273",
				Timeout: time.Duration(30) * time.Second,
				Account: Account{
					Username: "svr",
					Password: "svr",
				},
				Certificate: LClientCert{
					Cert: "./testcert/server.pem",
					Name: "error",
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
				{wantErr: false},
				{wantErr: true, errMsg: certErrMsg},
				{wantErr: true, errMsg: certErrMsg},
			},
		},
		{
			name: "Test 2 : Account error",
			ccfg: ClientConfig{
				Address: "0.0.0.0:8273",
				Timeout: time.Duration(10) * time.Second,
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
	option := &g.ServerOption{}
	opts := option.
		Create().
		CredsFromFile("./testcert/server.pem", "./testcert/server.key").
		Build()
	s := NewServer("svr",
		"svr",
		func(c context.Context, msg *Message) (*Message, error) {
			checkMsg(t, msg, msgCall)
			return msgCallResp, nil
		},
		func(stream Link_TalkServer) error {
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
	svr, err := g.NewServer(g.NetTCP, "0.0.0.0:8273", opts, func(svr *grpc.Server) {
		RegisterLinkServer(svr, s)
	})
	assert.NoError(t, err)
	defer svr.GracefulStop()

	wg := sync.WaitGroup{}
	for _, tt := range linkClientTests {
		t.Run(tt.name, func(t *testing.T) {
			option := &g.ClientOption{}
			opts := option.Create().
				CredsFromFile(tt.ccfg.Certificate.Cert, tt.ccfg.Certificate.Name).
				CustomCred(&g.CustomCred{
					Username: tt.ccfg.Account.Username,
					Password: tt.ccfg.Account.Password,
				}).Build()

			conn, err := g.NewClientConnect(tt.ccfg.Address, tt.ccfg.Timeout, opts)
			assert.Equal(t, tt.err[0].wantErr, err != nil)
			if conn != nil {
				cli := NewClient(conn)
				resp, err := cli.Call(tt.params.msgCall, tt.ccfg.Timeout)
				if tt.err[1].wantErr {
					assert.Error(t, err)
					assert.Equal(t, tt.err[1].errMsg, err.Error())
					return
				}
				assert.NoError(t, err)
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
				if tt.err[2].wantErr {
					assert.Error(t, err)
					assert.Equal(t, tt.err[2].errMsg, err.Error())
					return
				}
				assert.NoError(t, err)
				wg.Add(1)
				err = stream.CloseSend()
				assert.NoError(t, err)
				wg.Wait()
				err = conn.Close()
				assert.NoError(t, err)
			}
		})
	}
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
