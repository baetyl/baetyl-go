//go:build windows
// +build windows

package utils

import (
	"os"
	"syscall"
	"time"
	"unsafe"

	"github.com/baetyl/baetyl-go/v2/errors"
)

var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = modkernel32.NewProc("LockFileEx")
	procUnlockFileEx = modkernel32.NewProc("UnlockFileEx")
)

const (
	flagLockExclusive                     = 2
	flagLockFailImmediately               = 1
	errLockViolation        syscall.Errno = 0x21
	DefaultFlockRetry                     = time.Microsecond * 100
)

func Flock(file *os.File, timeout time.Duration) error {
	var t time.Time
	if timeout != 0 {
		t = time.Now()
	}
	fd := file.Fd()
	var flag uint32 = flagLockFailImmediately | flagLockExclusive
	var m1 uint32 = (1 << 32) - 1
	for {
		err := lockFileEx(syscall.Handle(fd), flag, 0, 1, 0, &syscall.Overlapped{
			Offset:     m1,
			OffsetHigh: m1,
		})
		if err == nil {
			return nil
		} else if err != errLockViolation {
			return err
		}
		if timeout != 0 && time.Since(t) > timeout-DefaultFlockRetry {
			return errors.New("timeout")
		}
		time.Sleep(DefaultFlockRetry)
	}
}

func Funlock(file *os.File) error {
	var m1 uint32 = (1 << 32) - 1
	return unlockFileEx(syscall.Handle(file.Fd()), 0, 1, 0, &syscall.Overlapped{
		Offset:     m1,
		OffsetHigh: m1,
	})
}

func lockFileEx(h syscall.Handle, flags, reserved, locklow, lockhigh uint32, ol *syscall.Overlapped) (err error) {
	r, _, err := procLockFileEx.Call(uintptr(h), uintptr(flags), uintptr(reserved), uintptr(locklow), uintptr(lockhigh), uintptr(unsafe.Pointer(ol)))
	if r == 0 {
		return err
	}
	return nil
}

func unlockFileEx(h syscall.Handle, reserved, locklow, lockhigh uint32, ol *syscall.Overlapped) (err error) {
	r, _, err := procUnlockFileEx.Call(uintptr(h), uintptr(reserved), uintptr(locklow), uintptr(lockhigh), uintptr(unsafe.Pointer(ol)), 0)
	if r == 0 {
		return err
	}
	return nil
}
