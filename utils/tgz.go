package utils

import (
	"compress/flate"

	"github.com/mholt/archiver"

	"github.com/baetyl/baetyl-go/v2/errors"
)

var defaultTgz = &archiver.TarGz{
	Tar: &archiver.Tar{
		MkdirAll:          true,
		OverwriteExisting: true,
	},
	CompressionLevel: flate.DefaultCompression,
}

// Tgz tar source files to destination file(.tgz/.tar.gz)
func Tgz(sources []string, destination string) error {
	return errors.Trace(defaultTgz.Archive(sources, destination))
}

// Untgz untar source file(.tgz/.tar.gz) to destination
func Untgz(source, destination string) error {
	return errors.Trace(defaultTgz.Unarchive(source, destination))
}
