package context

import (
	"os"

	"github.com/baetyl/baetyl-go/v2/errors"
)

// HostPathLib return HostPathLib
func HostPathLib() (string, error) {
	var hostPathLib string
	if val := os.Getenv(KeyBaetylHostPathLib); val == "" {
		err := os.Setenv(KeyBaetylHostPathLib, defaultHostPathLib)
		if err != nil {
			return "", errors.Trace(err)
		}
		hostPathLib = defaultHostPathLib
	} else {
		hostPathLib = val
	}
	return hostPathLib, nil
}

// RunMode return run mode of edge.
func RunMode() string {
	mode := os.Getenv(KeyRunMode)
	if mode != RunModeNative {
		mode = RunModeKube
	}
	return mode
}
