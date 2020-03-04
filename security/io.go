package security

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// CanReadCertAndKey returns true if the certificate and key files already exists,
// otherwise returns false. If lost one of cert and key, returns error.
func CanReadCertAndKey(certPath, keyPath string) (bool, error) {
	certReadable := canReadFile(certPath)
	keyReadable := canReadFile(keyPath)

	if certReadable == false && keyReadable == false {
		return false, nil
	}

	if certReadable == false {
		return false, fmt.Errorf("error reading %s, certificate and key must be supplied as a pair", certPath)
	}

	if keyReadable == false {
		return false, fmt.Errorf("error reading %s, certificate and key must be supplied as a pair", keyPath)
	}

	return true, nil
}

// If the file represented by path exists and
// readable, returns true otherwise returns false.
func canReadFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}

	defer f.Close()

	return true
}

// WriteCert writes the pem-encoded certificate data to certPath.
// The certificate file will be created with file mode 0644.
// If the certificate file already exists, it will be overwritten.
// The parent directory of the certPath will be created as needed with file mode 0755.
func WriteCert(data []byte, certPath string) error {
	if err := os.MkdirAll(filepath.Dir(certPath), os.FileMode(0755)); err != nil {
		return err
	}
	return ioutil.WriteFile(certPath, data, os.FileMode(0644))
}
