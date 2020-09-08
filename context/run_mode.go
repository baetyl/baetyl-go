package context

import (
	"os"
)

func RunMode() string {
	mode := os.Getenv(KeyRunMode)
	if mode != "native" {
		mode = "kube"
	}
	return mode
}
