package pki

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"io"
	"io/ioutil"
	"math/big"
	"strings"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/pki/models"
	"github.com/satori/go.uuid"
)

//go:generate mockgen -destination=../mock/pki/pki.go -package=pki github.com/baetyl/baetyl-go/v2/pki PKI

const (
	// TypeIssuingCA is a root certificate that can be used to issue sub-certificates
	TypeIssuingCA = "IssuingCA"
	// TypeIssuingSubCert is an issuing sub cert which is signed by issuing ca
	TypeIssuingSubCert = "IssuingSubCertificate"
)

type PKI interface {
	// GetRootCert certId: certificate ID
	GetRootCert(certId string) (*models.CertPem, error)
	// CreateRootCert info: request information for issuing a certificate;
	// durationDay: certificate validity period, in days; parentId: root ca certificate ID, used to issue sub-certificates
	CreateRootCert(info *x509.CertificateRequest, durationDay int, parentId string) (string, error)
	// CreateSelfSignedRootCert info: request information for issuing a certificate; durationDay: certificate validity period, in days;
	// generate a self-signed root certificate
	CreateSelfSignedRootCert(info *x509.CertificateRequest, durationDay int) (string, error)
	// DeleteRootCert rootId: certificate ID
	DeleteRootCert(rootId string) error

	// GetSubCert certId: certificate ID
	GetSubCert(certId string) ([]byte, error)
	// CreateSubCert csr: standard CSR request data; durationDay: certificate validity period, in days; rootId: root ca certificate ID
	CreateSubCert(csr []byte, durationDay int, rootId string) (string, error)
	// DeleteSubCert certId: certificate ID
	DeleteSubCert(certId string) error

	io.Closer
}

type defaultPKIClient struct {
	rootCaKey []byte
	rootCaCrt []byte
	sto       Storage
}

func NewPKIClient(keyFile, crtFile string, sto Storage) (PKI, error) {
	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, errors.Trace(err)
	}
	pem, err := ioutil.ReadFile(crtFile)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &defaultPKIClient{
		rootCaKey: key,
		rootCaCrt: pem,
		sto:       sto,
	}, nil
}

// root cert
func (p *defaultPKIClient) CreateRootCert(info *x509.CertificateRequest, durationDay int, parentId string) (string, error) {
	// get parent cert
	var caKeyByte []byte
	var caCrtByte []byte
	if len(parentId) == 0 {
		caKeyByte = p.rootCaKey
		caCrtByte = p.rootCaCrt
	} else {
		parentCert, err := p.sto.GetCert(parentId)
		if err != nil {
			return "", errors.Trace(err)
		}
		key, err := base64.StdEncoding.DecodeString(parentCert.PrivateKey)
		if err != nil {
			return "", errors.Trace(err)
		}
		caKeyByte = key
		content, err := base64.StdEncoding.DecodeString(parentCert.Content)
		if err != nil {
			return "", errors.Trace(err)
		}
		caCrtByte = content
	}

	// generate cert
	priv, err := GenCertPrivateKey(DefaultDSA, DefaultRSABits)
	if err != nil {
		return "", errors.Trace(err)
	}
	privByte, err := EncodeCertPrivateKey(priv)
	if err != nil {
		return "", errors.Trace(err)
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, info, priv.Key)
	if err != nil {
		return "", errors.Trace(err)
	}

	csrInfo, err := x509.ParseCertificateRequest(csr)
	if err != nil {
		return "", errors.Trace(err)
	}

	keyUsage := x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign

	// caInfo
	caKey, err := ParseCertPrivateKey(caKeyByte)
	if err != nil {
		return "", errors.Trace(err)
	}
	caCert, err := ParseCertificates(caCrtByte)
	if err != nil {
		return "", errors.Trace(err)
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
		SignatureAlgorithm:    SigAlgorithmType(caKey),
		KeyUsage:              keyUsage,
	}

	// The certificate is signed by parent. If parent is equal to template then the
	// certificate is self-signed. The parameter pub is the public key of the
	// signee and priv is the private key of the signer.
	cert, err := x509.CreateCertificate(rand.Reader, certInfo, caCert[0], csrInfo.PublicKey, caKey.Key)
	if err != nil {
		return "", errors.Trace(err)
	}

	// save cert
	certView := models.Cert{
		CertId:     strings.ReplaceAll(uuid.NewV4().String(), "-", ""),
		ParentId:   parentId,
		Type:       TypeIssuingCA,
		CommonName: info.Subject.CommonName,
		Csr:        base64.StdEncoding.EncodeToString([]byte(EncodeByteToPem(csr, CertificateRequestBlockType))),
		Content:    base64.StdEncoding.EncodeToString([]byte(EncodeByteToPem(cert, CertificateBlockType))),
		PrivateKey: base64.StdEncoding.EncodeToString(privByte),
		NotBefore:  certInfo.NotBefore,
		NotAfter:   certInfo.NotAfter,
	}
	err = p.sto.CreateCert(certView)
	if err != nil {
		return "", errors.Trace(err)
	}

	return certView.CertId, nil
}

func (p *defaultPKIClient) CreateSelfSignedRootCert(info *x509.CertificateRequest, durationDay int) (string, error) {
	// generate cert
	priv, err := GenCertPrivateKey(DefaultDSA, DefaultRSABits)
	if err != nil {
		return "", errors.Trace(err)
	}
	privByte, err := EncodeCertPrivateKey(priv)
	if err != nil {
		return "", errors.Trace(err)
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, info, priv.Key)
	if err != nil {
		return "", errors.Trace(err)
	}

	csrInfo, err := x509.ParseCertificateRequest(csr)
	if err != nil {
		return "", errors.Trace(err)
	}

	keyUsage := x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign

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
		SignatureAlgorithm:    SigAlgorithmType(priv),
		KeyUsage:              keyUsage,
	}

	// The certificate is signed by parent. If parent is equal to template then the
	// certificate is self-signed. The parameter pub is the public key of the
	// signee and priv is the private key of the signer.
	cert, err := x509.CreateCertificate(rand.Reader, certInfo, certInfo, csrInfo.PublicKey, priv.Key)
	if err != nil {
		return "", errors.Trace(err)
	}

	// save cert
	certView := models.Cert{
		CertId:     strings.ReplaceAll(uuid.NewV4().String(), "-", ""),
		Type:       TypeIssuingCA,
		CommonName: info.Subject.CommonName,
		Csr:        base64.StdEncoding.EncodeToString([]byte(EncodeByteToPem(csr, CertificateRequestBlockType))),
		Content:    base64.StdEncoding.EncodeToString([]byte(EncodeByteToPem(cert, CertificateBlockType))),
		PrivateKey: base64.StdEncoding.EncodeToString(privByte),
		NotBefore:  certInfo.NotBefore,
		NotAfter:   certInfo.NotAfter,
	}
	err = p.sto.CreateCert(certView)
	if err != nil {
		return "", errors.Trace(err)
	}

	return certView.CertId, nil
}

func (p *defaultPKIClient) GetRootCert(certId string) (*models.CertPem, error) {
	cert, err := p.sto.GetCert(certId)
	if err != nil {
		return nil, errors.Trace(err)
	}
	crt, err := base64.StdEncoding.DecodeString(cert.Content)
	if err != nil {
		return nil, errors.Trace(err)
	}
	key, err := base64.StdEncoding.DecodeString(cert.PrivateKey)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &models.CertPem{
		Crt: crt,
		Key: key,
	}, nil
}

func (p *defaultPKIClient) GetSubCert(certId string) ([]byte, error) {
	cert, err := p.sto.GetCert(certId)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return base64.StdEncoding.DecodeString(cert.Content)
}

func (p *defaultPKIClient) CreateSubCert(csr []byte, durationDay int, rootId string) (string, error) {
	// get ca cert
	ca, err := p.sto.GetCert(rootId)
	if err != nil {
		return "", errors.Trace(err)
	}
	if ca == nil {
		return "", errors.Trace(errors.Errorf("the root certificate(%s) not found", rootId))
	}

	priv, err := base64.StdEncoding.DecodeString(ca.PrivateKey)
	if err != nil {
		return "", errors.Trace(err)
	}
	content, err := base64.StdEncoding.DecodeString(ca.Content)
	if err != nil {
		return "", errors.Trace(err)
	}

	// parse ca cert
	caKey, err := ParseCertPrivateKey(priv)
	if err != nil {
		return "", errors.Trace(err)
	}
	caCert, err := ParseCertificates(content)
	if err != nil {
		return "", errors.Trace(err)
	}

	// create server data
	csrInfo, err := x509.ParseCertificateRequest(csr)
	if err != nil {
		return "", errors.Trace(err)
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

	certContent, err := x509.CreateCertificate(rand.Reader, certInfo, caCert[0], csrInfo.PublicKey, caKey.Key)
	if err != nil {
		return "", errors.Trace(err)
	}

	// save cert
	cert := models.Cert{
		CertId:     strings.ReplaceAll(uuid.NewV4().String(), "-", ""),
		ParentId:   rootId,
		Type:       TypeIssuingSubCert,
		CommonName: certInfo.Subject.CommonName,
		Csr:        base64.StdEncoding.EncodeToString([]byte(EncodeByteToPem(csr, CertificateRequestBlockType))),
		Content:    base64.StdEncoding.EncodeToString([]byte(EncodeByteToPem(certContent, CertificateBlockType))),
		NotBefore:  certInfo.NotBefore,
		NotAfter:   certInfo.NotAfter,
	}

	err = p.sto.CreateCert(cert)
	if err != nil {
		return "", errors.Trace(err)
	}

	return cert.CertId, nil
}

func (p *defaultPKIClient) DeleteRootCert(rootId string) error {
	count, err := p.sto.CountCertByParentId(rootId)
	if err != nil {
		return errors.Trace(err)
	}
	if count > 0 {
		return errors.Trace(errors.Errorf("the root certificate(%s) has been used by %d sub-certificate", rootId, count))
	}
	return p.sto.DeleteCert(rootId)
}

func (p *defaultPKIClient) DeleteSubCert(certId string) error {
	return p.sto.DeleteCert(certId)
}

func (p *defaultPKIClient) Close() error {
	return p.sto.Close()
}
