package utils

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTar(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "example")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	tmpfile, err := ioutil.TempFile(tmpdir, "test")
	assert.NoError(t, err)
	TarPath := tmpfile.Name() + ".tar"
	err = Tar([]string{tmpfile.Name()}, TarPath)
	assert.NoError(t, err)
	err = Untar(TarPath, tmpdir)
	assert.NoError(t, err)
	err = Untar(TarPath, tmpdir)
	assert.NoError(t, err)
}
