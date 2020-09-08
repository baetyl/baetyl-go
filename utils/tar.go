package utils

import (
	"github.com/mholt/archiver"

	"github.com/baetyl/baetyl-go/v2/errors"
)

var defaultTar = &archiver.Tar{
	MkdirAll:          true,
	OverwriteExisting: true,
}

// Tar tar source files to destination file
func Tar(sources []string, destination string) error {
	return errors.Trace(defaultTar.Archive(sources, destination))
}

// Untar untar source file to destination
func Untar(source, destination string) error {
	return errors.Trace(defaultTar.Unarchive(source, destination))
}
