package context

type SystemCert struct {
	ca  []byte
	crt []byte
	key []byte
}

// GetSystemCert get the signed certificate injected by the system, which can be used for TLS connection, etc.
func (s *SystemCert) GetSystemCert() (ca, crt, key []byte) {
	return s.ca, s.crt, s.key
}
