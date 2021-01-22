package utils

import (
	"os"
	"syscall"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
)

const (
	DefaultFlockRetry = time.Microsecond * 100
)

// only works on unix
func Flock(file *os.File, timeout time.Duration) error {
	var t time.Time
	if timeout != 0 {
		t = time.Now()
	}
	fd := file.Fd()
	flag := syscall.LOCK_NB | syscall.LOCK_EX
	for {
		err := syscall.Flock(int(fd), flag)
		if err == nil {
			return nil
		} else if err != syscall.EWOULDBLOCK {
			return err
		}
		if timeout != 0 && time.Since(t) > timeout-DefaultFlockRetry {
			return errors.New("timeout")
		}
		time.Sleep(DefaultFlockRetry)
	}
}

func Funlock(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
