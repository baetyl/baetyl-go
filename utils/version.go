package utils

import (
	"fmt"
	"runtime"
)

// Compile parameter
var (
	VERSION  string
	REVISION string
)

func Version() {
	fmt.Printf("Version:      %s\nGit revision: %s\nGo version:   %s\n", VERSION, REVISION, runtime.Version())
}
