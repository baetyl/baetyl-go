package context

type SystemResource struct {
	ca  []byte
	crt []byte
	key []byte
}

// GetSystemCert get the signed certificate injected by the system, which can be used for TLS connection, etc.
func (s *SystemResource) GetSystemCert() (ca, crt, key []byte) {
	return s.ca, s.crt, s.key
}
