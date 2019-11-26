package provisioners

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/OpenSource-THG/cfssl-issuer/provisioners/mock"

	api "github.com/OpenSource-THG/cfssl-issuer/api/v1beta1"
	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"k8s.io/apimachinery/pkg/types"
)

var validCABundle = readAndEncode("testdata/ca.pem")
var validCSR = readAndEncode("testdata/client.csr")

func TestProvisionerCreation(t *testing.T) {
	key := types.NamespacedName{
		Namespace: "default",
		Name:      "cfssl-issuer",
	}

	// first create the provisioner
	pro := newProvisioner(t, "http://test", "server")

	// now store it
	Store(key, pro)

	// now load it
	fetched, ok := Load(key)
	if !ok {
		t.Fatal("failed to retrieve provisioner")
	}

	if !reflect.DeepEqual(fetched, pro) {
		t.Fatal("returned object is not equal to expected object")
	}

	// create new spec and overwrite it
	newPro := newProvisioner(t, "http://test2", "client")
	Store(key, newPro)

	// now load it again
	fetched, ok = Load(key)
	if !ok {
		t.Fatal("failed to retrieve provisioner")
	}

	if !reflect.DeepEqual(fetched, newPro) {
		t.Fatal("returned object is not equal to expected object")
	}

	// new remove it
	Remove(key)

	// now load it again
	fetched, ok = Load(key)
	if ok || fetched != nil {
		t.Fatal("retrieved provisioner when it should have failed")
	}
}

func TestMultipleProvisioners(t *testing.T) {
	key1 := types.NamespacedName{
		Namespace: "default",
		Name:      "cfssl-issuer-1",
	}

	key2 := types.NamespacedName{
		Namespace: "default",
		Name:      "cfssl-issuer-2",
	}

	// first create the provisioner
	pro1 := newProvisioner(t, "http://test", "server")
	pro2 := newProvisioner(t, "http://test2", "server")

	// now store it
	Store(key1, pro1)
	Store(key2, pro2)

	// now load it
	fetched, ok := Load(key1)
	if !ok {
		t.Fatal("failed to retrieve provisioner")
	}

	if !reflect.DeepEqual(fetched, pro1) {
		t.Fatal("returned object is not equal to expected object")
	}

	fetched2, ok := Load(key2)
	if !ok {
		t.Fatal("failed to retrieve provisioner")
	}

	if !reflect.DeepEqual(fetched2, pro2) {
		t.Fatal("returned object is not equal to expected object")
	}
}

func TestProvisionerSigning(t *testing.T) {
	mockServer := mock.New()
	defer mockServer.Close()

	expectedCert, _ := ioutil.ReadFile("testdata/client.pem")
	expectedCA := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: mockServer.Certificate().Raw})

	csr := newCSR()
	pro := newProvisionerWithBundle(t, mockServer.URL, "client", encodeCert(mockServer.Certificate()))

	cert, ca, err := pro.Sign(context.Background(), csr)
	if err != nil {
		t.Fatalf("failed to sign csr: %v", err)
	}

	if !bytes.Equal(expectedCert, cert) {
		t.Error("returned cert does not matched expected value")
	}

	if !bytes.Equal(expectedCA, ca) {
		t.Error("returned ca does not matched expected value")
	}
}

//--- Helpers ---

func newProvisioner(t *testing.T, url, profile string) Provisioner {
	return newProvisionerWithBundle(t, url, profile, validCABundle)
}

func newProvisionerWithBundle(t *testing.T, url, profile string, bundle []byte) Provisioner {
	spec := &api.CfsslIssuerSpec{
		URL:      url,
		Profile:  profile,
		CABundle: bundle,
	}

	pro, err := New(spec)
	if err != nil {
		t.Fatalf("failed to create provisioner: %v", err)
	}

	return pro
}

func newCSR() *certmanager.CertificateRequest {
	return &certmanager.CertificateRequest{
		Spec: certmanager.CertificateRequestSpec{
			CSRPEM: validCSR,
		},
	}
}

func readAndEncode(f string) []byte {
	c, err := ioutil.ReadFile(f)
	if err != nil {
		panic("failed to read testdata")
	}

	return encode(c)
}

func encode(s []byte) []byte {
	if s == nil {
		return nil
	}

	r := base64.StdEncoding.EncodeToString(s)
	return []byte(r)
}

func encodeCert(c *x509.Certificate) []byte {
	b := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})

	r := base64.StdEncoding.EncodeToString(b)
	return []byte(r)
}
