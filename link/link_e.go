package link

import (
	"fmt"
	"sync"
	"time"
)

type pm struct {
	e *Event
	l time.Time // last send time
}

type publisher struct {
	d time.Duration
	m sync.Map
	c chan *pm
}

func newPublisher(dur time.Duration) *publisher {
	return &publisher{
		d: dur,
		c: make(chan *pm, 1),
	}
}

func (l *Linker) Send(src, dest string, content []byte) error {
	msg := packetMsg(src, dest, content)
	meta := &pm{
		e: NewEvent(msg),
		l: time.Now(),
	}
	err := l.stream.Send(msg)
	if err != nil {
		return err
	}
	if o, ok := l.publisher.m.LoadOrStore(msg.Context.ID, meta); ok {
		l.log.Error("message id conflict, to acknowledge old one")
		o.(*pm).e.Done()
	}
	select {
	case l.publisher.c <- meta:
		return nil
	case <-l.t.Dying():
		return ErrClientClosed
	}
}

func (l *Linker) republish(m *pm) error {
	m.l = time.Now()
	return l.stream.Send(m.e.msg)
}

func (l *Linker) waiting() error {
	l.log.Info("client starts to wait for message acknowledgement")

	var m *pm
	timer := time.NewTimer(l.publisher.d)
	defer timer.Stop()
	for {
		select {
		case m = <-l.publisher.c:
			for timer.Reset(l.publisher.d - time.Now().Sub(m.l)); !m.e.Wait(timer.C, l.t.Dying()); timer.Reset(l.publisher.d) {
				if err := l.republish(m); err != nil {
					return err
				}
			}
		case <-l.t.Dying():
			return nil
		}
	}
}

func (l *Linker) acknowledge(msg *Message) {
	id := msg.Context.ID
	m, ok := l.publisher.m.Load(id)
	if !ok {
		fmt.Printf("message id = %d is not found\n", id)
	}
	l.publisher.m.Delete(id)
	m.(*pm).e.Done()
}
