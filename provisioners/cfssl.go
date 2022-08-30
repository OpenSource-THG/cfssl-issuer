package provisioners

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	api "github.com/OpenSource-THG/cfssl-issuer/api/v1beta1"
	"github.com/cert-manager/cert-manager/pkg/util/pki"
	cfssl "github.com/cloudflare/cfssl/api/client"
	"k8s.io/apimachinery/pkg/types"
)

var (
	_ Provisioner = &CfsslProvisioner{}

	ErrInvalidBundle = errors.New("invalid ca bundle")

	p = new(sync.Map)
)

type Provisioner interface {
	Sign([]byte) ([]byte, []byte, error)
}

type certificateRequest struct {
	CSR     string `json:"certificate_request"`
	Profile string `json:"profile"`
}

type CfsslProvisioner struct {
	client  cfssl.Remote
	profile string
	ca      []byte
}

func New(spec api.CfsslIssuerSpec) (*CfsslProvisioner, error) {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	if ok := rootCAs.AppendCertsFromPEM(spec.CABundle); !ok {
		return nil, ErrInvalidBundle
	}

	tlsconfig := &tls.Config{
		RootCAs: rootCAs,
	}
	c := cfssl.NewServerTLS(spec.URL, tlsconfig)

	return &CfsslProvisioner{
		client:  c,
		profile: spec.Profile,
		ca:      spec.CABundle,
	}, nil
}

// Load returns a provisioner by NamespacedName.
func Load(namespacedName types.NamespacedName) (*CfsslProvisioner, bool) {
	v, ok := p.Load(namespacedName)
	if !ok {
		return nil, ok
	}
	p, ok := v.(*CfsslProvisioner)
	return p, ok
}

// Store adds a new provisioner to the collection by NamespacedName.
func Store(namespacedName types.NamespacedName, provisioner Provisioner) {
	p.Store(namespacedName, provisioner)
}

// Remove removes a provisioner from the collection
func Remove(namespacedName types.NamespacedName) {
	p.Delete(namespacedName)
}

func (cf *CfsslProvisioner) Sign(csrpem []byte) (resp, rootCA []byte, err error) {
	_, err = pki.DecodeX509CertificateRequestBytes(csrpem)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to validate CSR: %s", err)
	}

	csr := certificateRequest{
		CSR: string(csrpem),
	}
	if cf.profile != "" {
		csr.Profile = cf.profile
	}

	j, err := json.Marshal(csr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign certificate by cfssl: %s", err)
	}

	resp, err = cf.client.Sign(j)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign certificate by cfssl: %s", err)
	}

	// Decode CA chain and append all intermediate CAs to the response to be put in tls.crt
	caBundle, err := pki.DecodeX509CertificateChainBytes(cf.ca)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode CA chain: %s", err)
	}
	respCert, err := pki.DecodeX509CertificateBytes(resp)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode response cert: %s", err)
	}

	respChain := []*x509.Certificate{respCert}
	respChain = append(respChain, caBundle[:len(caBundle)-1]...)

	rootCA, err = pki.EncodeX509(caBundle[len(caBundle)-1])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode root CA: %s", err)
	}
	resp, err = pki.EncodeX509Chain(respChain)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode response cert chain: %s", err)
	}

	return resp, rootCA, nil
}
