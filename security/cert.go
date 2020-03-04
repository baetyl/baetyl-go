package security

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"time"
)

const (
	defaultKeySize              = 2048
	duration365d                = time.Hour * 24 * 365
	dataPath                    = "/var/lib/baetyl/certificates"
	ecPrivateKeyBlockType       = "EC PRIVATE KEY"
	rsaPrivateKeyBlockType      = "RSA PRIVATE KEY"
	privateKeyBlockType         = "PRIVATE KEY"
	certificateBlockType        = "CERTIFICATE"
	certificateRequestBlockType = "CERTIFICATE REQUEST"
)

// Config contains the basic fields required for creating a certificate
type Config struct {
	commonName   string
	organization []string
	altNames     altNames
	usages       []x509.ExtKeyUsage
}

type altNames struct {
	dnsNames []string
	ips      []net.IP
}

// NewRSAPrivateKey creates an RSA private key
func NewRSAPrivateKey(keySize int) (*rsa.PrivateKey, error) {
	if keySize == 0 {
		return rsa.GenerateKey(rand.Reader, defaultKeySize)
	}
	return rsa.GenerateKey(rand.Reader, keySize)
}

// NewSelfSignedCACert creates a CA certificate, the expired time is 10 years
func NewSelfSignedCACert(cfg Config, key crypto.Signer) (*x509.Certificate, error) {
	now := time.Now()
	tmpl := x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(0),
		Subject: pkix.Name{
			CommonName:   cfg.commonName,
			Organization: cfg.organization,
		},
		NotBefore:             now.UTC(),
		NotAfter:              now.Add(duration365d * 10).UTC(),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDERBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, key.Public(), key)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDERBytes)
}

// NewSignedCert creates a signed certificate using the given CA certificate and key
func NewSignedCert(csr *x509.CertificateRequest, key crypto.Signer, caCert *x509.Certificate, caKey crypto.Signer) (*x509.Certificate, error) {
	serial, err := genSerialNumber()
	if err != nil {
		return nil, err
	}

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   csr.Subject.CommonName,
			Organization: csr.Subject.Organization,
		},
		SerialNumber: serial,
		DNSNames:     csr.DNSNames,
		IPAddresses:  csr.IPAddresses,
		NotBefore:    caCert.NotBefore,
		NotAfter:     time.Now().Add(duration365d * 10).UTC(),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	}
	certDERBytes, err := x509.CreateCertificate(rand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDERBytes)
}

// EncodePrivateKeyPEM returns PEM-encoded private key data
func EncodePrivateKeyPEM(key *rsa.PrivateKey) []byte {
	block := pem.Block{
		Type:  rsaPrivateKeyBlockType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return pem.EncodeToMemory(&block)
}

// PrivateKey wraps a EC or RSA private key
type PrivateKey struct {
	Type string
	Key  interface{}
}

// Credentials holds a certificate, private key and trust chain
type Credentials struct {
	PrivateKey  *PrivateKey
	Certificate *x509.Certificate
}

// DecodePEMKey takes a key PEM byte array and returns a PrivateKey that represents
// Either an RSA or EC private key.
func DecodePEMKey(key []byte) (*PrivateKey, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, errors.New("key is not PEM encoded")
	}
	switch block.Type {
	case ecPrivateKeyBlockType:
		k, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return &PrivateKey{Type: ecPrivateKeyBlockType, Key: k}, nil
	case rsaPrivateKeyBlockType:
		k, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return &PrivateKey{Type: rsaPrivateKeyBlockType, Key: k}, nil
	default:
		return nil, fmt.Errorf("unsupported block type %s", block.Type)
	}
}

// DecodePEMCertificates takes a PEM encoded x509 certificates byte array and returns
// A x509 certificate and the block byte array.
func DecodePEMCertificates(crtb []byte) ([]*x509.Certificate, error) {
	certs := []*x509.Certificate{}
	for len(crtb) > 0 {
		var err error
		var cert *x509.Certificate

		cert, crtb, err = decodeCertificatePEM(crtb)
		if err != nil {
			return nil, err
		}
		if cert != nil {
			// it's a cert, add to pool
			certs = append(certs, cert)
		}
	}
	return certs, nil
}

func decodeCertificatePEM(crtb []byte) (*x509.Certificate, []byte, error) {
	block, crtb := pem.Decode(crtb)
	if block == nil {
		return nil, crtb, errors.New("invalid PEM certificate")
	}
	if block.Type != certificateBlockType {
		return nil, nil, nil
	}
	c, err := x509.ParseCertificate(block.Bytes)
	return c, crtb, err
}

// PEMCredentialsFromFiles takes a path for a key/cert pair and returns a validated Credentials wrapper with a trust chain.
func PEMCredentialsFromFiles(keyPath, certPath string) (*Credentials, error) {
	keyPem, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	certPem, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	pk, err := DecodePEMKey(keyPem)
	if err != nil {
		return nil, err
	}

	crts, err := DecodePEMCertificates(certPem)
	if err != nil {
		return nil, err
	}

	if len(crts) == 0 {
		return nil, errors.New("no certificates found")
	}

	match := matchCertificateAndKey(pk, crts[0])
	if !match {
		return nil, errors.New("error validating credentials: public and private key pair do not match")
	}

	creds := &Credentials{
		PrivateKey:  pk,
		Certificate: crts[0],
	}

	return creds, nil
}

func matchCertificateAndKey(pk *PrivateKey, cert *x509.Certificate) bool {
	switch pk.Type {
	case ecPrivateKeyBlockType:
		key := pk.Key.(*ecdsa.PrivateKey)
		pub, ok := cert.PublicKey.(*ecdsa.PublicKey)
		return ok && pub.X.Cmp(key.X) == 0 && pub.Y.Cmp(key.Y) == 0
	case rsaPrivateKeyBlockType:
		key := pk.Key.(*rsa.PrivateKey)
		pub, ok := cert.PublicKey.(*rsa.PublicKey)
		return ok && pub.N.Cmp(key.N) == 0 && pub.E == key.E
	default:
		return false
	}
}

// CertPoolFromPEM returns a CertPool from a PEM encoded certificates string.
func CertPoolFromPEM(certPem []byte) (*x509.CertPool, error) {
	certs, err := DecodePEMCertificates(certPem)
	if err != nil {
		return nil, err
	}
	if len(certs) == 0 {
		return nil, errors.New("no certificates found")
	}

	return certPoolFromCertificates(certs), nil
}

func certPoolFromCertificates(certs []*x509.Certificate) *x509.CertPool {
	pool := x509.NewCertPool()
	for _, c := range certs {
		pool.AddCert(c)
	}
	return pool
}

func genSerialNumber() (*big.Int, error) {
	serialNumLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNum, err := rand.Int(rand.Reader, serialNumLimit)
	if err != nil {
		return nil, fmt.Errorf("error generating serial number: %s", err)
	}
	return serialNum, nil
}
