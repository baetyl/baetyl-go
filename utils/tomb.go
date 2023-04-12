package utils

import (
	"sync"

	"github.com/pkg/errors"
	tb "gopkg.in/tomb.v2"
)

const (
	ini = int32(0)
	gos = int32(1)
)

// all errors
var (
	ErrStillAlive = tb.ErrStillAlive
	ErrDying      = tb.ErrDying
)

// Tomb wraps tomb.Tomb
type Tomb struct {
	t tb.Tomb
	s int32
	m sync.Mutex
}

// Go runs functions in new goroutines.
func (t *Tomb) Go(fs ...func() error) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = errors.Errorf("%v", p)
		}
	}()
	t.m.Lock()
	defer t.m.Unlock()
	t.s = gos
	for _, f := range fs {
		t.t.Go(f)
	}
	return
}

// Kill puts the tomb in a dying state for the given reason.
func (t *Tomb) Kill(reason error) {
	t.t.Kill(reason)
}

// Dying returns the channel that can be used to wait until
// t.Kill is called.
func (t *Tomb) Dying() <-chan struct{} {
	return t.t.Dying()
}

// Dead returns the channel that can be used to wait until all goroutines have finished running.
func (t *Tomb) Dead() <-chan struct{} {
	return t.t.Dead()
}

// Wait blocks until all goroutines have finished running, and
// then returns the reason for their death.
//
// If tomb does not start any goroutine, return quickly
func (t *Tomb) Wait() (err error) {
	t.m.Lock()
	if t.s == gos {
		err = t.t.Wait()
	}
	t.m.Unlock()
	return
}

// Alive returns true if the tomb is not in a dying or dead state.
func (t *Tomb) Alive() bool {
	return t.t.Alive()
}

// Err returns the death reason, or ErrStillAlive if the tomb is not in a dying or dead state.
func (t *Tomb) Err() (reason error) {
	return t.t.Err()
}
