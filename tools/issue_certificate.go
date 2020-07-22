package tools

import (
	"github.com/baetyl/baetyl-go/v2/pki/models"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	dbType     = "sqlite3"
	dbURL      = ":memory:"
	certTables = []string{`
CREATE TABLE baetyl_certificate
(
    cert_id          varchar(128)  PRIMARY KEY,
    parent_id        varchar(128)  NOT NULL DEFAULT '',
    type             varchar(64)   NOT NULL DEFAULT '',
    common_name      varchar(128)  NOT NULL DEFAULT '',
    description      varchar(256)  NOT NULL DEFAULT '',
    csr              varchar(2048) DEFAULT '',
    content          varchar(2048) DEFAULT '',
    private_key      varchar(2048) DEFAULT '',
    not_before       timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    not_after        timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_time      timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time      timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`}
)

type dbStorage struct {
	db *sqlx.DB
}

func newDBStorage() (*dbStorage, error) {
	db, err := sqlx.Open(dbType, dbURL)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	// create table
	for _, sql := range certTables {
		_, err := db.Exec(sql)
		if err != nil {
			return nil, err
		}
	}
	return &dbStorage{
		db: db,
	}, nil
}

func (d dbStorage) CreateCert(cert models.Cert) error {
	insertSQL := `
INSERT INTO baetyl_certificate (
cert_id, parent_id, type, common_name, 
description, csr, content, private_key, not_before, not_after) 
VALUES (?,?,?,?,?,?,?,?,?,?)
`
	_, err := d.db.Exec(insertSQL,
		cert.CertId, cert.ParentId, cert.Type,
		cert.CommonName, cert.Description, cert.Csr,
		cert.Content, cert.PrivateKey, cert.NotBefore, cert.NotAfter)
	return err
}

func (d dbStorage) DeleteCert(certId string) error {
	deleteSQL := `
DELETE FROM baetyl_certificate where cert_id=?
`
	_, err := d.db.Exec(deleteSQL, certId)
	return err
}

func (d dbStorage) UpdateCert(cert models.Cert) error {
	updateSQL := `
UPDATE baetyl_certificate SET parent_id=?,type=?,
common_name=?,description=?,csr=?,content=?,private_key=?,
not_before=?, not_after=? 
WHERE cert_id=?
`
	_, err := d.db.Exec(updateSQL,
		cert.ParentId, cert.Type, cert.CommonName, cert.Description, cert.Csr,
		cert.Content, cert.PrivateKey, cert.NotBefore, cert.NotAfter, cert.CertId)
	return err
}

func (d dbStorage) GetCert(certId string) (*models.Cert, error) {
	selectSQL := `
SELECT cert_id, parent_id, type, common_name, 
description, csr, content, private_key, not_before, not_after
FROM baetyl_certificate 
WHERE cert_id=? LIMIT 0,1
`
	var cert []models.Cert
	if err := d.db.Select(&cert, selectSQL, certId); err != nil {
		return nil, err
	}
	if len(cert) > 0 {
		return &cert[0], nil
	}
	return nil, nil
}

func (d dbStorage) CountCertByParentId(parentId string) (int, error) {
	selectSQL := `
SELECT count(cert_id) AS count 
FROM baetyl_certificate 
WHERE parent_id=?
`
	var res []struct {
		Count int `db:"count"`
	}
	if err := d.db.Select(&res, selectSQL, parentId); err != nil {
		return 0, err
	}
	return res[0].Count, nil
}

func (d dbStorage) Close() error {
	return d.db.Close()
}
