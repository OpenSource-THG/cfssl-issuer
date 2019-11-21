package controllers

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
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
		time.Sleep(time.Second * 8)

		fetched := &cfsslv1beta1.CfsslIssuer{}
		Eventually(func() bool {
			_ = k8sClient.Get(context.Background(), key, fetched)
			fmt.Printf("Fetched - %t: %v\n", fetched.IsReady(), fetched.Status)
			return fetched.IsReady()
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
})

func readAndEncode(f string) []byte {
	c, err := ioutil.ReadFile(f)
	if err != nil {
		log.Fatal("failed to read testdata")
	}

	r := base64.StdEncoding.EncodeToString(c)
	fmt.Println(r)
	return []byte(r)
}
