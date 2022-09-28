package utils

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTgz(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "example")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	tmpfile, err := ioutil.TempFile(tmpdir, "test")
	assert.NoError(t, err)
	TarPath := tmpfile.Name() + ".tgz"
	err = Tgz([]string{tmpfile.Name()}, TarPath)
	assert.NoError(t, err)
	err = Untgz(TarPath, tmpdir)
	assert.NoError(t, err)
	err = Untgz(TarPath, tmpdir)
	assert.NoError(t, err)
}

func TestTargz(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "example")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	tmpfile, err := ioutil.TempFile(tmpdir, "test")
	assert.NoError(t, err)
	TarPath := tmpfile.Name() + ".tar.gz"
	err = Tgz([]string{tmpfile.Name()}, TarPath)
	assert.NoError(t, err)
	err = Untgz(TarPath, tmpdir)
	assert.NoError(t, err)
	err = Untgz(TarPath, tmpdir)
	assert.NoError(t, err)
}
