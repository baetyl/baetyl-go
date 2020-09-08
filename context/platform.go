package context

import (
	"fmt"

	"github.com/containerd/containerd/platforms"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

type PlatformInfo = specs.Platform

func Platform() PlatformInfo {
	return platforms.DefaultSpec()
}

func PlatformString() string {
	pl := platforms.DefaultSpec()
	if pl.OS == "" {
		return "unknown"
	}
	if pl.Variant == "" {
		return fmt.Sprintf("%s-%s", pl.OS, pl.Architecture)
	}
	return fmt.Sprintf("%s-%s-%s", pl.OS, pl.Architecture, pl.Variant)
}
