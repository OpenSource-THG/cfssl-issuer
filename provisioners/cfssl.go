package provisioners

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	api "github.com/OpenSource-THG/cfssl-issuer/api/v1beta1"
	cfssl "github.com/cloudflare/cfssl/api/client"
	cfsslerr "github.com/cloudflare/cfssl/errors"
	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/jetstack/cert-manager/pkg/util/pki"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/types"
)

var (
	_ Provisioner = &cfsslProvisioner{}

	ErrInvalidBundle = errors.New("invalid ca bundle")

	p = new(sync.Map)
)

type Provisioner interface {
	Sign(context.Context, *certmanager.CertificateRequest) ([]byte, []byte, error)
}

type certificateRequest struct {
	CSR     string `json:"certificate_request"`
	Profile string `json:"profile"`
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

	if ok := rootCAs.AppendCertsFromPEM([]byte(spec.CABundle)); !ok {
		return nil, ErrInvalidBundle
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
	csrpem := cr.Spec.CSRPEM

	_, err := pki.DecodeX509CertificateRequestBytes(csrpem)
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
		return nil, nil, fmt.Errorf("failed to encode certificate request: %s", err)
	}

	t := prometheus.NewTimer(signRequests.WithLabelValues(cf.profile))
	resp, err := cf.client.Sign(j)
	t.ObserveDuration()
	if err != nil {
		signErrors.WithLabelValues(cf.profile).Inc()
		return nil, nil, fmt.Errorf("failed to sign certificate by cfssl: %w", err)
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

	rootCa, err := pki.EncodeX509(caBundle[len(caBundle)-1])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode root CA: %s", err)
	}
	resp, err = pki.EncodeX509Chain(respChain)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode response cert chain: %s", err)
	}

	return resp, rootCa, nil
}

// Retryable returns whether the given error from Sign is a transient
// error (e.g. due to the network).
func Retryable(err error) bool {
	var cerr *cfsslerr.Error
	if errors.As(err, &cerr) {
		category := cfsslerr.Category((cerr.ErrorCode / 1000) * 1000)
		reason := cfsslerr.Reason(cerr.ErrorCode % 1000)

		if category == cfsslerr.APIClientError &&
			reason == cfsslerr.ClientHTTPError &&
			strings.Contains(cerr.Message, "Request does not match policy whitelist") {
			return false
		}
	}
	// Conservatively assume everything else is transient.
	return true
}
