package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	cfsslv1beta1 "github.com/OpenSource-THG/cfssl-issuer/api/v1beta1"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("CertificateRequest Controller", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1

	const namespace = "default"

	It("Should ignore CertificateRequests configured for other issuers", func() {
		cleanup := setupCfsslIssuer(namespace, "cfssl-issuer-1")
		defer func() {
			_ = cleanup()
		}()

		tests := []struct {
			csr        *cmapi.CertificateRequest
			shouldPass bool
		}{
			{
				csr:        createCSR(namespace, "csr-use-cluster-issuer", "", "ClusterIssuer", "selfsigning-issuer"),
				shouldPass: false,
			},
			{
				csr:        createCSR(namespace, "csr-use-namespace-issuer", "", "Issuer", "selfsigning-issuer"),
				shouldPass: false,
			},
		}

		for _, tc := range tests {
			key := types.NamespacedName{
				Namespace: tc.csr.Namespace,
				Name:      tc.csr.Name,
			}
			// CSR should always be created successfully
			Expect(k8sClient.Create(context.Background(), tc.csr)).Should(Succeed())
			defer func() {
				_ = k8sClient.Delete(context.Background(), tc.csr)
			}()

			// If created for another issuer, we should do nothing
			// Else the conditions should change at some point
			a := Eventually(func() []cmapi.CertificateRequestCondition {
				f := cmapi.CertificateRequest{}
				_ = k8sClient.Get(context.Background(), key, &f)

				return f.Status.Conditions
			}, timeout, interval)

			switch {
			case tc.shouldPass:
				a.ShouldNot(BeNil())
			case !tc.shouldPass:
				a.Should(BeNil())
			}
		}
	})

})

func setupCfsslIssuer(namespace, name string) func() error {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	issuer := &cfsslv1beta1.CfsslIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: &cfsslv1beta1.CfsslIssuerSpec{
			URL:      "http://test",
			CABundle: caBundle,
		},
	}
	Expect(k8sClient.Create(context.Background(), issuer)).Should(Succeed())

	r := func() error {
		return k8sClient.Delete(context.Background(), issuer)
	}

	return r
}

func createCSR(namespace, name, group, kind, issuername string) *cmapi.CertificateRequest {
	return &cmapi.CertificateRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: cmapi.CertificateRequestSpec{
			IssuerRef: cmmeta.ObjectReference{
				Group: group,
				Kind:  kind,
				Name:  issuername,
			},
		},
	}
}
