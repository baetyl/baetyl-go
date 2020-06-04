package mqtt

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/baetyl/baetyl-go/errors"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
)

type stream struct {
	cli     *Client
	conn    Connection
	future  *Future
	tracker *Tracker
	tomb    utils.Tomb
	once    sync.Once
	mu      sync.Mutex
}

func (c *Client) connect() (*stream, error) {
	// dialing
	dialer := NewDialer(c.ops.TLSConfig, c.ops.Timeout)
	conn, err := dialer.Dial(c.ops.Address)
	if err != nil {
		return nil, errors.Trace(err)
	}

	// send connect
	connect := NewConnect()
	connect.ClientID = c.ops.ClientID
	connect.KeepAlive = uint16(math.Ceil(c.ops.KeepAlive.Seconds()))
	connect.CleanSession = c.ops.CleanSession
	connect.Username = c.ops.Username
	connect.Password = c.ops.Password
	// connect.Will = c.ops.WillMessage
	err = conn.Send(connect, false)
	if err != nil {
		conn.Close()
		return nil, errors.Trace(err)
	}

	s := &stream{
		cli:     c,
		conn:    conn,
		future:  NewFuture(),
		tracker: NewTracker(c.ops.KeepAlive),
	}
	s.tomb.Go(s.receiving)
	if c.ops.KeepAlive > 0 {
		s.tomb.Go(s.pinging)
	}
	err = s.future.Wait(c.ops.Timeout)
	if err != nil {
		s.close()
		return nil, errors.Trace(err)
	}
	return s, nil
}

func (s *stream) send(pkt Packet, async bool) error {
	s.tracker.Reset()

	s.mu.Lock()
	err := s.conn.Send(pkt, async)
	s.mu.Unlock()
	if err != nil {
		s.die("failed to send packet", err)
		return errors.Trace(err)
	}

	if ent := s.cli.log.Check(log.DebugLevel, "client sent a packet"); ent != nil {
		ent.Write(log.Any("pkt", fmt.Sprintf("%v", pkt)))
	}

	return nil
}

func (s *stream) sending(curr Packet) Packet {
	s.cli.log.Info("client starts to send packets")
	defer s.cli.log.Info("client has stopped sending packets")

	var err error
	if curr != nil {
		err = s.send(curr, true)
		if err != nil {
			return curr
		}
	}
	for {
		select {
		case pkt := <-s.cli.cache:
			err = s.send(pkt, true)
			if err != nil {
				return pkt
			}
		case <-s.cli.tomb.Dying():
			return nil
		case <-s.tomb.Dying():
			return nil
		}
	}
}

func (s *stream) receiving() error {
	s.cli.log.Info("client starts to receive packets")
	defer s.cli.log.Info("client has stopped receiving packets")

	var connacked bool
	for {
		pkt, err := s.conn.Receive()
		if err != nil {
			s.die("failed to receive packet", err)
			return errors.Trace(err)
		}

		if ent := s.cli.log.Check(log.DebugLevel, "client received a packet"); ent != nil {
			ent.Write(log.Any("pkt", fmt.Sprintf("%v", pkt)))
		}

		if !connacked {
			connacked = true
			err = s.cli.onConnack(pkt)
			if err != nil {
				s.die("failed to handle connack", err)
				return errors.Trace(err)
			}
			s.future.Complete(nil)
			continue
		}

		switch p := pkt.(type) {
		case *Publish:
			qos := p.Message.QOS
			uerr := s.cli.onPublish(p)
			if uerr != nil {
				s.cli.log.Warn("failed to handle publish packet in user code", log.Error(uerr))
			} else if !s.cli.ops.DisableAutoAck && qos == 1 {
				ack := NewPuback()
				ack.ID = p.ID
				err = s.send(ack, true)
			}
		case *Puback:
			err = s.cli.onPuback(p)
		case *Suback:
			err = s.cli.onSuback(p)
		case *Pingresp:
			s.tracker.Pong()
		case *Connack:
			err = errors.Trace(ErrClientAlreadyConnecting)
		default:
			err = errors.Errorf("packet (%v) not supported", p)
		}

		if err != nil {
			s.die("failed to handle packet", err)
			return errors.Trace(err)
		}
	}
}

func (s *stream) pinging() error {
	s.cli.log.Info("client starts to send pings")
	defer s.cli.log.Info("client has stopped sending pings")

	var err error
	var window time.Duration
	for {
		window = s.tracker.Window()
		if window < 0 {
			// check if a pong has already been sent
			if s.tracker.Pending() {
				s.die(ErrClientMissingPong.Error(), ErrClientMissingPong)
				return errors.Trace(ErrClientMissingPong)
			}

			s.tracker.Ping()
			err = s.send(NewPingreq(), false)
			if err != nil {
				return errors.Trace(err)
			}

			s.cli.log.Debug("client sent a ping")
		}

		select {
		case <-time.After(window):
			continue
		case <-s.cli.tomb.Dying():
			return nil
		case <-s.tomb.Dying():
			return nil
		}
	}
}

func (s *stream) die(msg string, err error) {
	s.once.Do(func() {
		s.future.Cancel(nil)
		s.tomb.Kill(err)
		if err == nil {
			s.send(NewDisconnect(), false)
		}
		s.cli.onError(msg, err)
	})
}

// ! called in the same goroutine with sending
func (s *stream) close() error {
	s.die("", nil)
	s.conn.Close()
	return errors.Trace(s.tomb.Wait())
}
