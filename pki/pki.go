package pki

import (
	"crypto/rand"
	"crypto/x509"
	"math/big"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
)

type CertPem struct {
	Crt []byte
	Key []byte
}

type PKI interface {
	// CreateSelfSignedRootCert info: request information for issuing a certificate;
	// durationDay: certificate validity period, in days;
	// generate a self-signed root certificate
	CreateSelfSignedRootCert(info *x509.CertificateRequest, durationDay int) (*CertPem, error)
	// CreateRootCert info: request information for issuing a certificate;
	// durationDay: certificate validity period, in days;
	// parent: root ca certificate, used to issue sub-certificates
	CreateRootCert(info *x509.CertificateRequest, durationDay int, parent *CertPem) (*CertPem, error)
	// CreateSubCert csr: standard CSR request data;
	// durationDay: certificate validity period, in days;
	// parent: root ca certificate, used to issue sub-certificates
	CreateSubCert(csr []byte, durationDay int, parent *CertPem) ([]byte, error)
	// CreateSubCertWithKey info: request information for issuing a certificate;
	// durationDay: certificate validity period, in days;
	// parent: root ca certificate, used to issue sub-certificates
	CreateSubCertWithKey(info *x509.CertificateRequest, durationDay int, parent *CertPem) (*CertPem, error)
}

type defaultPKIClient struct {
}

func NewPKIClient() (PKI, error) {
	return &defaultPKIClient{}, nil
}

func (p *defaultPKIClient) CreateSelfSignedRootCert(info *x509.CertificateRequest, durationDay int) (*CertPem, error) {
	return p.createRootCert(true, info, durationDay, nil)
}

func (p *defaultPKIClient) CreateRootCert(info *x509.CertificateRequest, durationDay int, parent *CertPem) (*CertPem, error) {
	return p.createRootCert(false, info, durationDay, parent)
}

func (p *defaultPKIClient) CreateSubCert(csr []byte, durationDay int, parent *CertPem) ([]byte, error) {
	// parse ca cert
	caKey, err := ParseCertPrivateKey(parent.Key)
	if err != nil {
		return nil, errors.Trace(err)
	}
	caCert, err := ParseCertificates(parent.Crt)
	if err != nil {
		return nil, errors.Trace(err)
	}

	// create server data
	csrInfo, err := x509.ParseCertificateRequest(csr)
	if err != nil {
		return nil, errors.Trace(err)
	}

	begin := time.Now()
	certInfo := &x509.Certificate{
		IsCA:                  false,
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               csrInfo.Subject,
		NotBefore:             begin,
		NotAfter:              begin.AddDate(0, 0, durationDay),
		EmailAddresses:        csrInfo.EmailAddresses,
		IPAddresses:           csrInfo.IPAddresses,
		URIs:                  csrInfo.URIs,
		DNSNames:              csrInfo.DNSNames,
		BasicConstraintsValid: true,
		SignatureAlgorithm:    SigAlgorithmType(caKey),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	cert, err := x509.CreateCertificate(rand.Reader, certInfo, caCert[0], csrInfo.PublicKey, caKey.Key)
	if err != nil {
		return nil, errors.Trace(err)
	}

	certInfo.Raw = cert
	crt, err := EncodeCertificates(certInfo)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return crt, nil
}

func (p *defaultPKIClient) CreateSubCertWithKey(info *x509.CertificateRequest, durationDay int, parent *CertPem) (*CertPem, error) {
	priv, err := GenCertPrivateKey(DefaultDSA, DefaultRSABits)
	if err != nil {
		return nil, errors.Trace(err)
	}
	privByte, err := EncodeCertPrivateKey(priv)
	if err != nil {
		return nil, errors.Trace(err)
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, info, priv.Key)
	if err != nil {
		return nil, errors.Trace(err)
	}
	crt, err := p.CreateSubCert(csr, durationDay, parent)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &CertPem{
		Crt: crt,
		Key: privByte,
	}, nil
}

func (p *defaultPKIClient) createRootCert(isSelfSigned bool, info *x509.CertificateRequest, durationDay int, parent *CertPem) (*CertPem, error) {
	priv, err := GenCertPrivateKey(DefaultDSA, DefaultRSABits)
	if err != nil {
		return nil, errors.Trace(err)
	}
	privByte, err := EncodeCertPrivateKey(priv)
	if err != nil {
		return nil, errors.Trace(err)
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, info, priv.Key)
	if err != nil {
		return nil, errors.Trace(err)
	}

	csrInfo, err := x509.ParseCertificateRequest(csr)
	if err != nil {
		return nil, errors.Trace(err)
	}

	begin := time.Now()
	certInfo := &x509.Certificate{
		IsCA:                  true,
		Subject:               info.Subject,
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		NotBefore:             begin,
		NotAfter:              begin.AddDate(0, 0, durationDay),
		EmailAddresses:        info.EmailAddresses,
		IPAddresses:           info.IPAddresses,
		URIs:                  info.URIs,
		DNSNames:              info.DNSNames,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}

	var caKey *PrivateKey
	var caCert *x509.Certificate
	if isSelfSigned {
		caKey = priv
		certInfo.SignatureAlgorithm = SigAlgorithmType(caKey)

		caCert = certInfo
	} else {
		caKey, err = ParseCertPrivateKey(parent.Key)
		if err != nil {
			return nil, errors.Trace(err)
		}
		certInfo.SignatureAlgorithm = SigAlgorithmType(caKey)

		caCerts, err := ParseCertificates(parent.Crt)
		if err != nil {
			return nil, errors.Trace(err)
		}
		caCert = caCerts[0]
	}

	// The certificate is signed by parent. If parent is equal to template then the
	// certificate is self-signed. The parameter pub is the public key of the
	// signee and priv is the private key of the signer.
	cert, err := x509.CreateCertificate(rand.Reader, certInfo, caCert, csrInfo.PublicKey, caKey.Key)
	if err != nil {
		return nil, errors.Trace(err)
	}

	certInfo.Raw = cert
	crt, err := EncodeCertificates(certInfo)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &CertPem{
		Crt: crt,
		Key: privByte,
	}, nil
}
