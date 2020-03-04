package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
)

// GenerateCSR creates a X.509 certificate sign request and private key.
func GenerateCSR(cfg Config, key interface{}) ([]byte, error) {
	template, err := genCSRTemplate(cfg)
	if err != nil {
		return nil, fmt.Errorf("error generating csr template: %s", err)
	}
	template.SignatureAlgorithm = sigType(key)
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, template, key)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSR: %s", err)
	}

	csrPemBlock := &pem.Block{
		Type:  certificateRequestBlockType,
		Bytes: csrDER,
	}
	return pem.EncodeToMemory(csrPemBlock), nil
}

func genCSRTemplate(cfg Config) (*x509.CertificateRequest, error) {
	return &x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: cfg.organization,
			CommonName:   cfg.commonName,
		},
		DNSNames:    cfg.altNames.dnsNames,
		IPAddresses: cfg.altNames.ips,
	}, nil
}

// ParsePemCSR constructs a x509 Certificate Request using the
// given PEM-encoded certificate signing request.
func ParsePemCSR(csrPem []byte) (*x509.CertificateRequest, error) {
	block, _ := pem.Decode(csrPem)
	if block == nil {
		return nil, fmt.Errorf("certificate signing request is not properly encoded")
	}
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse X.509 certificate signing request: %s", err)
	}
	return csr, nil
}

func sigType(privateKey interface{}) x509.SignatureAlgorithm {
	// customize the signature for RSA keys, depending on the key size
	if privateKey, ok := privateKey.(*rsa.PrivateKey); ok {
		keySize := privateKey.N.BitLen()
		switch {
		case keySize >= 4096:
			return x509.SHA512WithRSA
		case keySize >= 3072:
			return x509.SHA384WithRSA
		default:
			return x509.SHA256WithRSA
		}
	}
	return x509.UnknownSignatureAlgorithm
}
