package utils

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTomb(t *testing.T) {
	tb := new(Tomb)
	err := tb.Go(func() error {
		<-tb.Dying()
		return nil
	})
	assert.NoError(t, err)
	assert.EqualError(t, tb.Err(), ErrStillAlive.Error())
	tb.Kill(nil)
	err = tb.Wait()
	assert.NoError(t, err)

	tb = new(Tomb)
	err = tb.Go(func() error {
		<-tb.Dying()
		return fmt.Errorf("abc")
	})
	assert.NoError(t, err)
	tb.Kill(nil)
	tb.Kill(nil)
	err = tb.Wait()
	assert.EqualError(t, err, "abc")
	err = tb.Wait()
	assert.EqualError(t, err, "abc")

	tb = new(Tomb)
	tb.Kill(fmt.Errorf("abc"))
	err = tb.Go(func() error {
		<-tb.Dying()
		return nil
	})
	assert.NoError(t, err)
	err = tb.Wait()
	assert.EqualError(t, err, "abc")

	tb = new(Tomb)
	tb.Kill(fmt.Errorf("abc"))
	err = tb.Wait()
	assert.NoError(t, err)
	err = tb.Go(func() error {
		<-tb.Dying()
		return nil
	})
	assert.NoError(t, err)
	tb.Kill(nil)
	err = tb.Wait()
	assert.EqualError(t, err, "abc")

	tb = new(Tomb)
	tb.Kill(nil)
	err = tb.Wait()
	assert.NoError(t, err)
	err = tb.Go(func() error {
		<-tb.Dying()
		return nil
	})
	assert.NoError(t, err)
	tb.Kill(fmt.Errorf("abc"))
	tb.Kill(fmt.Errorf("efd"))
	err = tb.Wait()
	assert.EqualError(t, err, "abc")
	assert.EqualError(t, tb.Err(), "abc")

	tb = new(Tomb)
	err = tb.Go(func() error {
		<-tb.Dying()
		return nil
	})
	assert.NoError(t, err)
	tb.Kill(nil)
	err = tb.Wait()
	assert.NoError(t, err)
	err = tb.Go(func() error {
		<-tb.Dying()
		return nil
	})
	assert.EqualError(t, err, "tomb.Go called after all goroutines terminated")

	tb = new(Tomb)
	err = tb.Go(func() error {
		return nil
	})
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	err = tb.Go(func() error {
		return nil
	})
	assert.EqualError(t, err, "tomb.Go called after all goroutines terminated")
	tb.Kill(nil)
	err = tb.Wait()
	assert.NoError(t, err)
}

func TestKillErrStillAlivePanic(t *testing.T) {
	tb := new(Tomb)
	defer func() {
		err := recover()
		if err != "tomb: Kill with ErrStillAlive" {
			t.Fatalf("Wrong panic on Kill(ErrStillAlive): %v", err)
		}
		checkState(t, tb, false, false, ErrStillAlive)
	}()
	b := tb.Alive()
	assert.Equal(t, true, b)
	assert.EqualError(t, tb.Err(), ErrStillAlive.Error())
	tb.Kill(ErrStillAlive)
}

func checkState(t *testing.T, tm *Tomb, wantDying, wantDead bool, wantErr error) {
	select {
	case <-tm.Dying():
		if !wantDying {
			t.Error("<-Dying: should block")
		}
	default:
		if wantDying {
			t.Error("<-Dying: should not block")
		}
	}
	seemsDead := false
	select {
	case <-tm.Dead():
		if !wantDead {
			t.Error("<-Dead: should block")
		}
		seemsDead = true
	default:
		if wantDead {
			t.Error("<-Dead: should not block")
		}
	}
	if err := tm.Err(); err != wantErr {
		t.Errorf("Err: want %#v, got %#v", wantErr, err)
	}
	if wantDead && seemsDead {
		waitErr := tm.Wait()
		switch {
		case waitErr == ErrStillAlive:
			t.Errorf("Wait should not return ErrStillAlive")
		case !reflect.DeepEqual(waitErr, wantErr):
			t.Errorf("Wait: want %#v, got %#v", wantErr, waitErr)
		}
	}
}

func BenchmarkA(b *testing.B) {
	msg := "aaa"
	msgchan := make(chan string, b.N)
	var tomb Tomb
	for i := 0; i < b.N; i++ {
		select {
		case <-tomb.Dying():
			continue
		case msgchan <- msg:
		default: // discard if channel is full
		}
	}
}

func BenchmarkB(b *testing.B) {
	msg := "aaa"
	msgchan := make(chan string, b.N)
	var tomb Tomb
	for i := 0; i < b.N; i++ {
		if !tomb.Alive() {
			continue
		}
		select {
		case msgchan <- msg:
		default: // discard if channel is full
		}
	}
}
