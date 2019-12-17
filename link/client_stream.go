package link

import (
	fmt "fmt"
	io "io"
	"sync"

	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// all flags
const (
	FlagRetain = 0x1
	FlagAck    = 0x2
)

// stream stream stream
type stream struct {
	grpc.ClientStream
	cfg ClientConfig
	obs Observer
	log *log.Logger
	utils.Tomb
	sync.Once
}

// NewClient creates a new stream of client
func newStream(ctx context.Context, cli *Client) (*stream, error) {
	cs, err := cli.cli.Talk(ctx, grpc.WaitForReady(true))
	if err != nil {
		return nil, err
	}
	s := &stream{
		ClientStream: cs,
		cfg:          cli.cfg,
		obs:          cli.obs,
		log:          log.With(log.Any("link", "stream")),
	}
	s.Go(s.receiving)
	return s, nil
}

// Send sends a message
func (s *stream) Send(msg *Message) (err error) {
	err = s.SendMsg(msg)
	if err != nil {
		s.die(err)
	}
	return
}

// Close closes stream
func (s *stream) Close() error {
	if !s.Alive() {
		return nil
	}
	s.Kill(nil)
	s.close(nil)
	return s.Wait()
}

// closes stream by itself
func (s *stream) die(err error) {
	if !s.Alive() {
		return
	}
	s.Kill(err)
	go s.close(err)
}

func (s *stream) close(err error) {
	s.Do(func() {
		s.log.Info("stream is closing", log.Error(err))
		defer s.log.Info("stream has closed", log.Error(err))
		if err != nil && err != io.EOF {
			s.onErr(err)
		}
		s.CloseSend()
		s.Wait()
	})
}

func (s *stream) receiving() error {
	s.log.Info("stream starts to receive messages")
	defer s.log.Info("stream has stopped receiving messages")

	for {
		msg := new(Message)
		err := s.RecvMsg(msg)
		if err != nil {
			s.die(err)
			return err
		}

		if ent := s.log.Check(log.DebugLevel, "stream received a message"); ent != nil {
			ent.Write(log.Any("msg", fmt.Sprintf("%v", msg)))
		}

		err = s.onMsg(msg)
		if err != nil {
			s.die(err)
			return err
		}
	}
}

func (s *stream) onMsg(msg *Message) error {
	if s.obs == nil {
		return nil
	}
	if msg.Ack() {
		return s.obs.OnAck(msg)
	}
	err := s.obs.OnMsg(msg)
	if err != nil {
		return err
	}
	if !s.cfg.DisableAutoAck {
		ack := &Message{}
		ack.Context.ID = msg.Context.ID
		ack.Context.Flags = FlagAck
		err = s.Send(ack)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *stream) onErr(err error) {
	if s.obs == nil {
		return
	}
	s.obs.OnErr(err)
}
