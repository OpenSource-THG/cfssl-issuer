package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	cfsslv1beta1 "github.com/OpenSource-THG/cfssl-issuer/api/v1beta1"
	cmutil "github.com/cert-manager/cert-manager/pkg/api/util"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const namespace = "default"

var _ = Describe("CertificateRequest Controller", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1

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
				csr:        createCSR("csr-use-cluster-issuer", "", "ClusterIssuer", "selfsigning-issuer"),
				shouldPass: false,
			},
			{
				csr:        createCSR("csr-use-namespace-issuer", "", "Issuer", "selfsigning-issuer"),
				shouldPass: false,
			},
			{
				csr:        createCSR("csr-use-cfssl-namespace", "certmanager.thg.io", "CfsslIssuer", "cfssl-issuer-1"),
				shouldPass: true,
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

	It("Should mark certificate request as pending when using namespace scoped issuer that doesn't exist", func() {
		csr := createCSR("csr-pending", "certmanager.thg.io", "CfsslIssuer", "cfssl-issuer-pending")
		key := types.NamespacedName{
			Namespace: csr.Namespace,
			Name:      csr.Name,
		}
		// CSR should always be created successfully
		Expect(k8sClient.Create(context.Background(), csr)).Should(Succeed())
		time.Sleep(time.Second * 2)
		defer func() {
			_ = k8sClient.Delete(context.Background(), csr)
		}()

		Eventually(func() bool {
			f := &cmapi.CertificateRequest{}
			err := k8sClient.Get(context.Background(), key, f)
			if err != nil {
				return false
			}

			return cmutil.CertificateRequestHasCondition(f, cmapi.CertificateRequestCondition{
				Type:   cmapi.CertificateRequestConditionReady,
				Status: cmmeta.ConditionFalse,
				Reason: cmapi.CertificateRequestReasonPending,
			})
		}, timeout, interval).Should(BeTrue())

	})

	It("Should mark certificate request as pending when using a cluster scoped issuer that doesn't exist", func() {
		csr := createCSR("csr-pending", "certmanager.thg.io", "CfsslClusterIssuer", "cfssl-issuer-pending")
		key := types.NamespacedName{
			Namespace: csr.Namespace,
			Name:      csr.Name,
		}
		// CSR should always be created successfully
		Expect(k8sClient.Create(context.Background(), csr)).Should(Succeed())
		time.Sleep(time.Second * 2)
		defer func() {
			_ = k8sClient.Delete(context.Background(), csr)
		}()

		Eventually(func() bool {
			f := &cmapi.CertificateRequest{}
			err := k8sClient.Get(context.Background(), key, f)
			if err != nil {
				return false
			}

			return cmutil.CertificateRequestHasCondition(f, cmapi.CertificateRequestCondition{
				Type:   cmapi.CertificateRequestConditionReady,
				Status: cmmeta.ConditionFalse,
				Reason: cmapi.CertificateRequestReasonPending,
			})
		}, timeout, interval).Should(BeTrue())

	})

	It("Should mark certificate request as ready when using a namespace scoped issuer", func() {
		issuerKey := types.NamespacedName{
			Name:      "cfssl-issuer-ready",
			Namespace: namespace,
		}
		issuer := &cfsslv1beta1.CfsslIssuer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      issuerKey.Name,
				Namespace: issuerKey.Namespace,
			},
			Spec: cfsslv1beta1.CfsslIssuerSpec{
				URL:      mockCfsslServer.URL,
				CABundle: encodeCert(mockCfsslServer.Certificate()),
			},
		}

		Expect(k8sClient.Create(context.Background(), issuer)).Should(Succeed())
		time.Sleep(time.Second * 2)
		defer func() {
			_ = k8sClient.Delete(context.Background(), issuer)
		}()

		csr := createCSR("csr-ready", "certmanager.thg.io", "CfsslIssuer", "cfssl-issuer-ready")
		key := types.NamespacedName{
			Namespace: csr.Namespace,
			Name:      csr.Name,
		}

		// CSR should always be created successfully
		Expect(k8sClient.Create(context.Background(), csr)).Should(Succeed())
		time.Sleep(time.Second * 2)
		defer func() {
			_ = k8sClient.Delete(context.Background(), csr)
		}()

		Eventually(func() bool {
			f := &cmapi.CertificateRequest{}
			err := k8sClient.Get(context.Background(), key, f)
			if err != nil {
				return false
			}

			return cmutil.CertificateRequestHasCondition(f, cmapi.CertificateRequestCondition{
				Type:   cmapi.CertificateRequestConditionReady,
				Status: cmmeta.ConditionTrue,
				Reason: cmapi.CertificateRequestReasonIssued,
			})
		}, timeout, interval).Should(BeTrue())

	})

	It("Should mark certificate request as pending when referencing a deleted issuer", func() {
		issuerKey := types.NamespacedName{
			Name:      "cfssl-issuer-deleted",
			Namespace: namespace,
		}
		issuer := &cfsslv1beta1.CfsslIssuer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      issuerKey.Name,
				Namespace: issuerKey.Namespace,
			},
			Spec: cfsslv1beta1.CfsslIssuerSpec{
				URL:      mockCfsslServer.URL,
				CABundle: encodeCert(mockCfsslServer.Certificate()),
			},
		}

		Expect(k8sClient.Create(context.Background(), issuer)).Should(Succeed())
		time.Sleep(time.Second * 2)

		Expect(k8sClient.Delete(context.Background(), issuer)).Should(Succeed())
		time.Sleep(time.Second * 2)

		csr := createCSR("csr-ready", "certmanager.thg.io", "CfsslIssuer", "cfssl-issuer-deleted")
		key := types.NamespacedName{
			Namespace: csr.Namespace,
			Name:      csr.Name,
		}

		// CSR should always be created successfully
		Expect(k8sClient.Create(context.Background(), csr)).Should(Succeed())
		time.Sleep(time.Second * 2)
		defer func() {
			_ = k8sClient.Delete(context.Background(), csr)
		}()

		Eventually(func() bool {
			f := &cmapi.CertificateRequest{}
			err := k8sClient.Get(context.Background(), key, f)
			if err != nil {
				return false
			}

			return cmutil.CertificateRequestHasCondition(f, cmapi.CertificateRequestCondition{
				Type:   cmapi.CertificateRequestConditionReady,
				Status: cmmeta.ConditionFalse,
				Reason: cmapi.CertificateRequestReasonPending,
			})
		}, timeout, interval).Should(BeTrue())

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
		Spec: cfsslv1beta1.CfsslIssuerSpec{
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

func createCSR(name, group, kind, issuername string) *cmapi.CertificateRequest {
	csrblock := readAndEncode("testdata/client.csr")

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
			Request: csrblock,
		},
	}
}
