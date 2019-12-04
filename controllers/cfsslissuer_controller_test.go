package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	cfsslv1beta1 "github.com/OpenSource-THG/cfssl-issuer/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var caBundle = readAndEncode("testdata/ca.pem")

var _ = Describe("CfsslIssuer Controller", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1

	const namespace = "default"

	It("Should handle scope correctly", func() {
		key := types.NamespacedName{
			Name:      "cfssl-issuer-1",
			Namespace: "default",
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
		time.Sleep(time.Second * 2)

		fetched := &cfsslv1beta1.CfsslIssuer{}
		Eventually(func() bool {
			_ = k8sClient.Get(context.Background(), key, fetched)
			return fetched.IsReady()
		}, timeout, interval).Should(BeTrue())

		By("Updating the scope")
		fetched.Spec.URL = "http://test.new.url"

		Expect(k8sClient.Update(context.Background(), fetched)).Should(Succeed())
		time.Sleep(time.Second * 2)
		Eventually(func() bool {
			f := &cfsslv1beta1.CfsslIssuer{}
			_ = k8sClient.Get(context.Background(), key, f)
			return f.IsReady()
		}, timeout, interval).Should(BeTrue())

		By("Deleting the scope")
		Eventually(func() error {
			f := &cfsslv1beta1.CfsslIssuer{}
			_ = k8sClient.Get(context.Background(), key, f)
			return k8sClient.Delete(context.Background(), issuer)
		}).Should(Succeed())

		Eventually(func() error {
			f := &cfsslv1beta1.CfsslIssuer{}
			return k8sClient.Get(context.Background(), key, f)
		}).ShouldNot(Succeed())
	})

	It("Should validate params", func() {
		By("Requiring CABundle")
		missingBundle := &cfsslv1beta1.CfsslIssuer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cfssl-issuer-missing-bundle",
				Namespace: namespace,
			},
			Spec: &cfsslv1beta1.CfsslIssuerSpec{
				URL: "http://test",
			},
		}

		Expect(k8sClient.Create(context.Background(), missingBundle)).ShouldNot(Succeed())
		defer func() {
			_ = k8sClient.Delete(context.Background(), missingBundle)
		}()

		By("Requiring validCABundle")
		invalidBundleKey := types.NamespacedName{
			Name:      "cfssl-issuer-invalid-bundle",
			Namespace: namespace,
		}
		invalidBundle := &cfsslv1beta1.CfsslIssuer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      invalidBundleKey.Name,
				Namespace: namespace,
			},
			Spec: &cfsslv1beta1.CfsslIssuerSpec{
				URL:      "http://test",
				CABundle: []byte("this-isnt-base64"),
			},
		}

		Expect(k8sClient.Create(context.Background(), invalidBundle)).Should(Succeed())
		time.Sleep(time.Second * 2)

		Eventually(func() bool {
			f := &cfsslv1beta1.CfsslIssuer{}
			err := k8sClient.Get(context.Background(), invalidBundleKey, f)
			if err != nil || f.Status == nil {
				return false
			}

			for _, cond := range f.Status.Conditions {
				if cond.Type != cfsslv1beta1.ConditionReady {
					continue
				}

				if cond.Status == cfsslv1beta1.ConditionFalse &&
					cond.Reason == "Error" &&
					cond.Message == "failed to initialize provisioner" {
					return true
				}
			}

			return false
		}).Should(BeTrue())
		defer func() {
			_ = k8sClient.Delete(context.Background(), invalidBundle)
		}()

		By("Requiring URL")
		missingURLKey := types.NamespacedName{
			Name:      "cfssl-issuer-missing-url",
			Namespace: namespace,
		}
		missingURL := &cfsslv1beta1.CfsslIssuer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      missingURLKey.Name,
				Namespace: missingURLKey.Namespace,
			},
			Spec: &cfsslv1beta1.CfsslIssuerSpec{
				CABundle: caBundle,
			},
		}

		Expect(k8sClient.Create(context.Background(), missingURL)).Should(Succeed())
		time.Sleep(time.Second * 2)

		Eventually(func() bool {
			f := &cfsslv1beta1.CfsslIssuer{}
			err := k8sClient.Get(context.Background(), missingURLKey, f)
			if err != nil || f.Status == nil {
				return false
			}

			for _, cond := range f.Status.Conditions {
				if cond.Type != cfsslv1beta1.ConditionReady {
					continue
				}

				if cond.Status == cfsslv1beta1.ConditionFalse && cond.Reason == "Validation" {
					return true
				}
			}

			return false
		}).Should(BeTrue())
		defer func() {
			_ = k8sClient.Delete(context.Background(), missingURL)
		}()
	})
})
