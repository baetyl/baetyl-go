package mqtt

import (
	"fmt"
	"math"
	"sync"
	"time"

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
}

func (c *Client) connect() (*stream, error) {
	// dialing
	dialer := NewDialer(c.tls, c.cfg.Timeout)
	conn, err := dialer.Dial(c.cfg.Address)
	if err != nil {
		return nil, err
	}

	// send connect
	connect := NewConnect()
	connect.ClientID = c.cfg.ClientID
	connect.KeepAlive = uint16(math.Ceil(c.cfg.KeepAlive.Seconds()))
	connect.CleanSession = c.cfg.CleanSession
	connect.Username = c.cfg.Username
	connect.Password = c.cfg.Password
	// connect.Will = c.cfg.WillMessage
	err = conn.Send(connect, false)
	if err != nil {
		conn.Close()
		return nil, err
	}

	s := &stream{
		cli:     c,
		conn:    Connection{conn},
		future:  NewFuture(),
		tracker: NewTracker(c.cfg.KeepAlive),
	}
	s.tomb.Go(s.receiving)
	if c.cfg.KeepAlive > 0 {
		s.tomb.Go(s.pinging)
	}
	err = s.future.Wait(c.cfg.Timeout)
	if err != nil {
		s.close()
		return nil, err
	}
	return s, nil
}

func (s *stream) send(pkt Packet, async bool) error {
	s.tracker.Reset()

	err := s.conn.Send(pkt, async)
	if err != nil {
		s.die("failed to send packet", err)
		return err
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
			return err
		}

		if ent := s.cli.log.Check(log.DebugLevel, "client received a packet"); ent != nil {
			ent.Write(log.Any("pkt", fmt.Sprintf("%v", pkt)))
		}

		if !connacked {
			connacked = true
			err = s.cli.onConnack(pkt)
			if err != nil {
				s.die("failed to handle connack", err)
				return err
			}
			s.future.Complete()
			continue
		}

		switch p := pkt.(type) {
		case *Publish:
			err = s.cli.onPublish(p)
		case *Puback:
			err = s.cli.onPuback(p)
		case *Suback:
			err = s.cli.onSuback(p)
		case *Pingresp:
			s.tracker.Pong()
		case *Connack:
			err = ErrClientAlreadyConnecting
		default:
			err = fmt.Errorf("packet (%v) not supported", p)
		}

		if err != nil {
			s.die("failed to handle packet", err)
			return err
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
				return ErrClientMissingPong
			}

			s.tracker.Ping()
			err = s.send(NewPingreq(), false)
			if err != nil {
				return err
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
		s.future.Cancel()
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
	return s.tomb.Wait()
}
