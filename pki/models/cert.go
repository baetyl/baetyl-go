package models

import (
	"time"
)

type Cert struct {
	CertId      string    `db:"cert_id"`
	ParentId    string    `db:"parent_id"`
	Type        string    `db:"type"`
	CommonName  string    `db:"common_name"`
	Csr         string    `db:"csr"`         // base64
	Content     string    `db:"content"`     // base64
	PrivateKey  string    `db:"private_key"` // base64
	Description string    `db:"description"`
	NotBefore   time.Time `db:"not_before"`
	NotAfter    time.Time `db:"not_after"`
}

type CertPem struct {
	Crt []byte
	Key []byte
}
