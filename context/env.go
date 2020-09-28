package context

import (
	"os"

	"github.com/baetyl/baetyl-go/v2/errors"
)

func HostPathLib() (string, error) {
	var hostPathLib string
	if val := os.Getenv(KeyBaetylHostPathLib); val == "" {
		err := os.Setenv(KeyBaetylHostPathLib, DefaultHostPathLib)
		if err != nil {
			return "", errors.Trace(err)
		}
		hostPathLib = DefaultHostPathLib
	} else {
		hostPathLib = val
	}
	return hostPathLib, nil
}
