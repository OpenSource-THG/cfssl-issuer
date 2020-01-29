module github.com/OpenSource-THG/cfssl-issuer

go 1.13

replace (
	k8s.io/api => k8s.io/api v0.0.0-20191114100237-2cd11237263f
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191114105316-e8706470940d
	k8s.io/apimachinery => k8s.io/apimachinery v0.15.10-beta.0
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191114102923-bf973bc1a46c
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191114101336-8cba805ad12d
	k8s.io/code-generator => k8s.io/code-generator v0.15.10-beta.0
	k8s.io/component-base => k8s.io/component-base v0.0.0-20191114102239-843ff05e8ff4
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191114103707-3917fe134eab
)

require (
	cloud.google.com/go v0.52.0 // indirect
	github.com/cloudflare/cfssl v1.4.1
	github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jetstack/cert-manager v0.12.0
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.4.0 // indirect
	github.com/stretchr/testify v1.4.0
	go.uber.org/atomic v1.5.1 // indirect
	go.uber.org/multierr v1.4.0 // indirect
	go.uber.org/zap v1.13.0 // indirect
	golang.org/x/crypto v0.0.0-20200128174031-69ecbb4d6d5d // indirect
	golang.org/x/sys v0.0.0-20200124204421-9fbb57f87de9 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/tools v0.0.0-20200129045341-207d3de1faaf // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
	k8s.io/api v0.0.0-20191114100352-16d7abae0d2a
	k8s.io/apimachinery v0.0.0-20191028221656-72ed19daf4bb
	k8s.io/client-go v0.0.0-20191114101535-6c5935290e33
	k8s.io/utils v0.0.0-20190801114015-581e00157fb1
	sigs.k8s.io/controller-runtime v0.3.1-0.20191022174215-ad57a976ffa1
)
