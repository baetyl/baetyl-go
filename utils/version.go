package utils

import (
	"fmt"
)

// Compile parameter
var (
	VERSION  = "unknown"
	REVISION = "unknown"
)

func Version() string {
	return fmt.Sprintf(" Version: %s\nRevision: %s", VERSION, REVISION)
}

func PrintVersion() {
	fmt.Println(Version())
}
