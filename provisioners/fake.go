package provisioners

import (
	"context"

	api "github.com/OpenSource-THG/cfssl-issuer/api/v1beta1"
	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
)

var _ Provisioner = &fakeProvisioner{}

type fakeProvisioner struct {
	issuer *api.CfsslIssuer
}

func NewFake(i *api.CfsslIssuer) *fakeProvisioner {
	return &fakeProvisioner{
		issuer: i,
	}
}

func (cf *fakeProvisioner) Sign(ctx context.Context, cr *certmanager.CertificateRequest) ([]byte, []byte, error) {
	return nil, nil, nil
}
