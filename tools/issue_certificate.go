package main

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/pki"
)

type helper struct {
	cli pki.PKI
}

type AltNames struct {
	DNSNames []string   `json:"dnsNames,omitempty"`
	IPs      []net.IP   `json:"ips,omitempty"`
	Emails   []string   `json:"emails,omitempty"`
	URIs     []*url.URL `json:"uris,omitempty"`
}

func genCsr(cn string, alt AltNames) *x509.CertificateRequest {
	return &x509.CertificateRequest{
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Organization:       []string{"Linux Foundation Edge"},
			OrganizationalUnit: []string{"BAETYL"},
			Locality:           []string{"Haidian District"},
			Province:           []string{"Beijing"},
			StreetAddress:      []string{"Baidu Campus"},
			PostalCode:         []string{"100093"},
			CommonName:         cn,
		},
		DNSNames:       alt.DNSNames,
		EmailAddresses: alt.Emails,
		IPAddresses:    alt.IPs,
		URIs:           alt.URIs,
	}
}

func (h *helper) createRoot() (*pki.CertPem, error) {
	cn := "root.ca"
	csrInfo := genCsr(cn, AltNames{
		IPs: []net.IP{
			net.IPv4(0, 0, 0, 0),
			net.IPv4(127, 0, 0, 1),
		},
		URIs: []*url.URL{
			{
				Scheme: "https",
				Host:   "localhost",
			},
		},
	})
	cert, err := h.cli.CreateSelfSignedRootCert(csrInfo, 50*365)
	if err != nil {
		return nil, errors.Trace(err)
	}
	fmt.Println("ca.crt")
	fmt.Println(string(cert.Crt))
	fmt.Println("ca.key")
	fmt.Println(string(cert.Key))
	return cert, nil
}

func (h *helper) createSub(cn string, alt AltNames, parent *pki.CertPem) (*pki.CertPem, error) {
	cert, err := h.cli.CreateSubCertWithKey(genCsr(cn, alt), 20*365, parent)
	if err != nil {
		return nil, errors.Trace(err)
	}
	fmt.Println("sub.crt")
	fmt.Println(string(cert.Crt))
	fmt.Println("sub.key")
	fmt.Println(string(cert.Key))
	return cert, nil
}

func (h *helper) issueCert() error {
	ca, err := h.createRoot()
	if err != nil {
		return errors.Trace(err)
	}
	err = ioutil.WriteFile("output/ca.crt", ca.Crt, 0666)
	if err != nil {
		return errors.Trace(err)
	}
	err = ioutil.WriteFile("output/ca.key", ca.Key, 0666)
	if err != nil {
		return errors.Trace(err)
	}

	client, err := h.createSub("client", AltNames{
		IPs: []net.IP{
			net.IPv4(0, 0, 0, 0),
			net.IPv4(127, 0, 0, 1),
		},
		URIs: []*url.URL{
			{
				Scheme: "https",
				Host:   "localhost",
			},
		},
	}, ca)
	if err != nil {
		return errors.Trace(err)
	}
	err = ioutil.WriteFile("output/client.crt", client.Crt, 0666)
	if err != nil {
		return errors.Trace(err)
	}
	err = ioutil.WriteFile("output/client.key", client.Key, 0666)
	if err != nil {
		return errors.Trace(err)
	}

	server, err := h.createSub("server", AltNames{
		IPs: []net.IP{
			net.IPv4(0, 0, 0, 0),
			net.IPv4(127, 0, 0, 1),
		},
		URIs: []*url.URL{
			{
				Scheme: "https",
				Host:   "localhost",
			},
		},
	}, ca)
	if err != nil {
		return errors.Trace(err)
	}
	err = ioutil.WriteFile("output/server.crt", server.Crt, 0666)
	if err != nil {
		return errors.Trace(err)
	}
	err = ioutil.WriteFile("output/server.key", server.Key, 0666)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func main() {
	cli, err := pki.NewPKIClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	h := helper{
		cli: cli,
	}
	err = h.issueCert()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
