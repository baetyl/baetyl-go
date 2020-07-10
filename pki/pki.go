package pki

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"github.com/baetyl/baetyl-go/errors"
	"io"
	"io/ioutil"
	"math/big"
	"strings"
	"time"

	"github.com/baetyl/baetyl-go/pki/models"
	"github.com/satori/go.uuid"
)

//go:generate mockgen -destination=../mock/pki/pki.go -package=pki github.com/baetyl/baetyl-go/pki PKI

const (
	TypeIssuingCA = "IssuingCA"
	// TypeIssuingSubCert is an issuing sub cert which is signed by issuing ca
	TypeIssuingSubCert = "IssuingSubCertificate"

	// 证书有效期，以天为单位 [1, 50*365]
	DefaultCADuration      = 50 * 365
	DefaultSubCertDuration = 20 * 365
)

type PKI interface {
	CreateRootCert(info *x509.CertificateRequest, parentId string) (string, error)
	GetCert(certId string) ([]byte, error)
	CreateSubCert(csr []byte, rootId string) (string, error)
	DeleteRootCert(rootId string) error
	DeleteSubCert(certId string) error
	io.Closer
}

type defaultPKIClient struct {
	rootCaKey []byte
	rootCaCrt []byte
	pvc       PVC
}

func NewPKIClient(keyFile, crtFile string, pvc PVC) (PKI, error) {
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
		pvc:       pvc,
	}, nil
}

// root cert
func (p *defaultPKIClient) CreateRootCert(info *x509.CertificateRequest, parentId string) (string, error) {
	// get parent cert
	var caKeyByte []byte
	var caCrtByte []byte
	if len(parentId) == 0 {
		caKeyByte = p.rootCaKey
		caCrtByte = p.rootCaCrt
	} else {
		parentCert, err := p.pvc.GetCert(parentId)
		if err != nil {
			return "", errors.Trace(err)
		}
		priv, err := GenCertPrivateKey(DefaultDSA, DefaultRSABits)
		if err != nil {
			return "", errors.Trace(err)
		}
		caKeyByte, err = EncodeCertPrivateKey(priv)
		if err != nil {
			return "", errors.Trace(err)
		}
		content, err := base64.StdEncoding.DecodeString(parentCert.Content)
		if err != nil {
			return "", errors.Trace(err)
		}
		caCrtByte = content
	}

	// generate cert
	caKey, err := ParseCertPrivateKey(caKeyByte)
	if err != nil {
		return "", errors.Trace(err)
	}
	caCert, err := ParseCertificates(caCrtByte)
	if err != nil {
		return "", errors.Trace(err)
	}

	csr, err := x509.CreateCertificateRequest(rand.Reader, info, caKey.Key)
	if err != nil {
		return "", errors.Trace(err)
	}

	csrInfo, err := x509.ParseCertificateRequest(csr)
	if err != nil {
		return "", errors.Trace(err)
	}

	keyUsage := x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign

	certInfo := &x509.Certificate{
		IsCA:                  true,
		Subject:               info.Subject,
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().AddDate(0, 0, DefaultCADuration).UTC(),
		EmailAddresses:        info.EmailAddresses,
		IPAddresses:           info.IPAddresses,
		URIs:                  info.URIs,
		DNSNames:              info.DNSNames,
		BasicConstraintsValid: true,
		SignatureAlgorithm:    SigAlgorithmType(caKey),
		KeyUsage:              keyUsage,
	}

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
		PrivateKey: base64.StdEncoding.EncodeToString(caKeyByte),
		NotBefore:  certInfo.NotBefore,
		NotAfter:   certInfo.NotAfter,
	}
	err = p.pvc.CreateCert(certView)
	if err != nil {
		return "", errors.Trace(err)
	}

	return certView.CertId, nil
}

func (p *defaultPKIClient) GetCert(certId string) ([]byte, error) {
	cert, err := p.pvc.GetCert(certId)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return base64.StdEncoding.DecodeString(cert.Content)
}

func (p *defaultPKIClient) CreateSubCert(csr []byte, rootId string) (string, error) {
	// get ca cert
	ca, err := p.pvc.GetCert(rootId)
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

	certInfo := &x509.Certificate{
		IsCA:                  false,
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               csrInfo.Subject,
		NotBefore:             time.Now().UTC(),
		NotAfter:              caCert[0].NotAfter,
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

	err = p.pvc.CreateCert(cert)
	if err != nil {
		return "", errors.Trace(err)
	}

	return cert.CertId, nil
}

func (p *defaultPKIClient) DeleteRootCert(rootId string) error {
	count, err := p.pvc.CountCertByParentId(rootId)
	if err != nil {
		return errors.Trace(err)
	}
	if count > 0 {
		return errors.Trace(errors.Errorf("the root certificate(%s) has been used by %d sub-certificate", rootId, count))
	}
	return p.pvc.DeleteCert(rootId)
}

func (p *defaultPKIClient) DeleteSubCert(certId string) error {
	return p.pvc.DeleteCert(certId)
}

func (p *defaultPKIClient) Close() error {
	return p.pvc.Close()
}
