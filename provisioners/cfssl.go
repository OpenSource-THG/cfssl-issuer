package provisioners

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"

	"github.com/jetstack/cert-manager/pkg/util/pki"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/OpenSource-THG/cfssl-issuer/api/v1beta1"
	cfssl "github.com/cloudflare/cfssl/api/client"
	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
)

var _ Provisioner = &cfsslProvisioner{}

var p = new(sync.Map)

type Provisioner interface {
	Sign(context.Context, *certmanager.CertificateRequest) ([]byte, []byte, error)
}

type cfsslProvisioner struct {
	client  cfssl.Remote
	profile string
	ca      []byte
}

func New(spec *api.CfsslIssuerSpec) (*cfsslProvisioner, error) {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	caBundle, err := base64.StdEncoding.DecodeString(string(spec.CABundle))
	if err != nil {
		return nil, fmt.Errorf("unable to decode ca bundle: %w", err)
	}

	if ok := rootCAs.AppendCertsFromPEM([]byte(caBundle)); !ok {
		return nil, errors.New("invalid ca bundle")
	}

	tlsconfig := &tls.Config{
		RootCAs: rootCAs,
	}
	c := cfssl.NewServerTLS(spec.URL, tlsconfig)

	return &cfsslProvisioner{
		client:  c,
		profile: spec.Profile,
		ca:      spec.CABundle,
	}, nil
}

// Load returns a provisioner by NamespacedName.
func Load(namespacedName types.NamespacedName) (*cfsslProvisioner, bool) {
	v, ok := p.Load(namespacedName)
	if !ok {
		return nil, ok
	}
	p, ok := v.(*cfsslProvisioner)
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

func (cf *cfsslProvisioner) Sign(ctx context.Context, cr *certmanager.CertificateRequest) ([]byte, []byte, error) {
	csrpem, err := base64.StdEncoding.DecodeString(string(cr.Spec.CSRPEM))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode CSR: %s", err)
	}

	_, err = pki.DecodeX509CertificateRequestBytes(csrpem)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to validate CSR: %s", err)
	}

	resp, err := cf.client.Sign(cr.Spec.CSRPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign certificate by cfssl: %s", err)
	}

	return resp, cf.ca, nil
}
