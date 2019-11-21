package provisioners

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"sync"

	"github.com/jetstack/cert-manager/pkg/util/pki"
	"k8s.io/apimachinery/pkg/types"

	cfssl "github.com/cloudflare/cfssl/api/client"
	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	api "github.com/OpenSource-THG/cfssl-issuer/api/v1beta1"
)

var _ Provisioner = &cfsslProvisioner{}

var clusterProvisioners = new(sync.Map)
var provisioners = new(sync.Map)

type Provisioner interface {
	Sign(context.Context, *certmanager.CertificateRequest) ([]byte, []byte, error)
}

type cfsslProvisioner struct {
	client  cfssl.Remote
	profile string
	ca      []byte
}

func New(i *api.CfsslIssuer) (*cfsslProvisioner, error) {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	if ok := rootCAs.AppendCertsFromPEM(i.Spec.CABundle); !ok {
		return nil, errors.New("invalid ca bundle")
	}

	tlsconfig := &tls.Config{
		RootCAs: rootCAs,
	}
	c := cfssl.NewServerTLS(i.Spec.URL, tlsconfig)

	return &cfsslProvisioner{
		client:  c,
		profile: i.Spec.Profile,
		ca:      i.Spec.CABundle,
	}, nil
}

// LoadCluster returns a provisioner by Name.
func LoadCluster(name string) (*cfsslProvisioner, bool) {
	v, ok := clusterProvisioners.Load(name)
	if !ok {
		return nil, ok
	}
	p, ok := v.(*cfsslProvisioner)
	return p, ok
}

// StoreCluster adds a new provisioner to the collection by Name.
func StoreCluster(name string, provisioner Provisioner) {
	provisioners.Store(name, provisioner)
}

// Load returns a provisioner by NamespacedName.
func Load(namespacedName types.NamespacedName) (*cfsslProvisioner, bool) {
	v, ok := provisioners.Load(namespacedName)
	if !ok {
		return nil, ok
	}
	p, ok := v.(*cfsslProvisioner)
	return p, ok
}

// Store adds a new provisioner to the collection by NamespacedName.
func Store(namespacedName types.NamespacedName, provisioner Provisioner) {
	provisioners.Store(namespacedName, provisioner)
}

func (cf *cfsslProvisioner) Sign(ctx context.Context, cr *certmanager.CertificateRequest) ([]byte, []byte, error) {
	_, err := pki.DecodeX509CertificateRequestBytes(cr.Spec.CSRPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode CSR for signing: %s", err)
	}

	resp, err := cf.client.Sign(cr.Spec.CSRPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign certificate by cfssl: %s", err)
	}

	return resp, cf.ca, nil
}
