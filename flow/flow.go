// Package flow can be used to test MQTT packet flows.
package flow

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// A Conn defines an abstract interface for connections used with a Flow.
type Conn interface {
	Send(pkt interface{}) error
	Receive() (interface{}, error)
	Close() error
}

// The Pipe pipes packets from Send to Receive.
type Pipe struct {
	pipe  chan interface{}
	close chan struct{}
}

// NewPipe returns a new Pipe.
func NewPipe() *Pipe {
	return &Pipe{
		pipe:  make(chan interface{}),
		close: make(chan struct{}),
	}
}

// Send returns packet on next Receive call.
func (conn *Pipe) Send(pkt interface{}) error {
	select {
	case conn.pipe <- pkt:
		return nil
	case <-conn.close:
		return errors.New("closed")
	}
}

// Receive returns the packet being sent with Send.
func (conn *Pipe) Receive() (interface{}, error) {
	select {
	case pkt := <-conn.pipe:
		return pkt, nil
	case <-conn.close:
		return nil, io.EOF
	}
}

// Close will close the conn and let Send and Receive return errors.
func (conn *Pipe) Close() error {
	close(conn.close)
	return nil
}

// All available action types.
const (
	actionSend byte = iota
	actionReceive
	actionSkip
	actionRun
	actionClose
	actionEnd
)

// An Action is a step in a flow.
type action struct {
	kind     byte
	packets  []interface{}
	fn       func()
	ch       chan struct{}
	duration time.Duration
}

// A Flow is a sequence of actions that can be tested against a connection.
type Flow struct {
	debug   bool
	actions []action
}

// New returns a new flow.
func New() *Flow {
	return &Flow{
		actions: make([]action, 0),
	}
}

// Debug will activate the debug mode.
func (f *Flow) Debug() *Flow {
	f.debug = true
	return f
}

// Send will send the specified packets.
func (f *Flow) Send(pkts ...interface{}) *Flow {
	f.add(action{
		kind:    actionSend,
		packets: pkts,
	})

	return f
}

// Receive will receive and match the specified packets out of order.
func (f *Flow) Receive(pkts ...interface{}) *Flow {
	f.add(action{
		kind:    actionReceive,
		packets: pkts,
	})

	return f
}

// Skip will receive the specified packets without matching out of order.
func (f *Flow) Skip(pkts ...interface{}) *Flow {
	f.add(action{
		kind:    actionSkip,
		packets: pkts,
	})

	return f
}

// Run will call the supplied function and wait until it returns.
func (f *Flow) Run(fn func()) *Flow {
	f.add(action{
		kind: actionRun,
		fn:   fn,
	})

	return f
}

// Close will immediately close the connection.
func (f *Flow) Close() *Flow {
	f.add(action{
		kind: actionClose,
	})

	return f
}

// End will match proper connection close.
func (f *Flow) End() *Flow {
	f.add(action{
		kind: actionEnd,
	})

	return f
}

// Test starts the flow on the given Conn and reports to the specified test.
func (f *Flow) Test(conn Conn) error {
	// handle all actions
	for _, action := range f.actions {
		err := f.test(conn, action)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *Flow) test(conn Conn, action action) error {
	switch action.kind {
	case actionSend:
		// send all saved packets
		for _, pkt := range action.packets {
			if f.debug {
				fmt.Printf("sending packet: %v...\n", pkt)
			}

			// send a single packet
			err := conn.Send(pkt)
			if err != nil {
				return fmt.Errorf("error sending packet: %v", err)
			}
		}
	case actionReceive:
		// initialize store
		store := make(map[int]string)
		for i, pkt := range action.packets {
			store[i] = fmt.Sprintf("%v", pkt)
		}

	receive:
		// keep going until we have all packets
		for len(store) > 0 {
			if f.debug {
				fmt.Printf("receiving packet...\n")
			}

			// wait for next packet
			pkt, err := conn.Receive()
			if err != nil {
				return fmt.Errorf("expected to receive a packet but got error: %v", err)
			}

			if f.debug {
				fmt.Println("received packet:", pkt)
			}

			// check packet
			for i, p := range store {
				if p == fmt.Sprintf("%v", pkt) {
					delete(store, i)
					continue receive
				}
			}

			return fmt.Errorf("unexpected packet %q", pkt)
		}
	case actionSkip:
		// initialize store
		store := make(map[int]string)
		for i, pkt := range action.packets {
			store[i] = fmt.Sprintf("%T", pkt)
		}

	skip:
		// keep going until we have all packets
		for len(store) > 0 {
			if f.debug {
				fmt.Printf("skiping packet...\n")
			}

			// wait for next packet
			pkt, err := conn.Receive()
			if err != nil {
				return fmt.Errorf("expected to skip over a received packet but got error: %v", err)
			}

			if f.debug {
				fmt.Printf("received packet: %T", pkt)
			}

			// check packet
			for i, p := range store {
				if p == fmt.Sprintf("%T", pkt) {
					delete(store, i)
					continue skip
				}
			}

			return fmt.Errorf("unexpected to receive a packet of type %T", pkt)
		}
	case actionRun:
		if f.debug {
			fmt.Printf("running...\n")
		}

		// run function
		action.fn()
	case actionClose:
		if f.debug {
			fmt.Printf("closing...\n")
		}

		// close connection
		err := conn.Close()
		if err != nil {
			return fmt.Errorf("expected connection to close successfully but got error: %v", err)
		}
	case actionEnd:
		if f.debug {
			fmt.Printf("ending...\n")
		}

		// wait for end of file
		pkt, err := conn.Receive()
		if err != nil && !strings.Contains(err.Error(), "EOF") {
			return fmt.Errorf("expected EOF but got %v", err)
		}
		if pkt != nil {
			return fmt.Errorf("expected no packet but got %v", pkt)
		}
	}

	return nil
}

// TestAsync starts the flow on the given Conn and reports to the specified test
// asynchronously.
func (f *Flow) TestAsync(conn Conn, timeout time.Duration) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		select {
		case <-time.After(timeout):
			errCh <- errors.New("timed out waiting for flow to complete")
		case errCh <- f.Test(conn):
		}
	}()

	return errCh
}

// add will add the specified action.
func (f *Flow) add(action action) {
	f.actions = append(f.actions, action)
}
