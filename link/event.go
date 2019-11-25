package link

import (
	"time"
)

type acknowledge struct {
	done chan struct{}
}

// Done acknowledges once after event is handled
func (a *acknowledge) _done(id uint64) {
	close(a.done)
}

// Wait waits until acknowledged or cancelled
func (a *acknowledge) _wait(timeout <-chan time.Time, cancel <-chan struct{}) bool {
	if a.done == nil {
		return true
	}
	select {
	case <-a.done:
		return true
	case <-timeout:
		return false
	case <-cancel:
		return false
	}
}

// Event event with message and acknowledge
type Event struct {
	msg *Message
	ack *acknowledge
}

// Done the event is acknowledged
func (e *Event) Done() {
	if e.ack != nil {
		e.ack._done(e.msg.Context.ID)
	}
}

// Wait waits until acknowledged (returns true), cancelled or timed out
func (e *Event) Wait(timeout <-chan time.Time, cancel <-chan struct{}) bool {
	return e.ack._wait(timeout, cancel)
}

// NewEvent creates a new event
func NewEvent(msg *Message) *Event {
	return &Event{
		msg: msg,
		ack: &acknowledge{
			done: make(chan struct{}),
		},
	}
}
