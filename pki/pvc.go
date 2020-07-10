package pki

import (
	"io"

	"github.com/baetyl/baetyl-go/pki/models"
)

//go:generate mockgen -destination=../mock/pki/pvc.go -package=pki github.com/baetyl/baetyl-go/pki PVC

type PVC interface {
	CreateCert(cert models.Cert) error
	DeleteCert(certId string) error
	UpdateCert(cert models.Cert) error
	GetCert(certId string) (*models.Cert, error)
	CountCertByParentId(parentId string) (int, error)
	io.Closer
}
