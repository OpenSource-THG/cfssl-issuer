module github.com/opensource-thg/cfssl-issuer

go 1.13

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190918155943-95b840bb6a1f
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190918161926-8f644eb6e783
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190918160949-bfa5e2e684ad
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190912054826-cd179ad6a269
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190918160511-547f6c5d7090
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190918161219-8c8f079fddc3
)

require (
	github.com/cloudflare/cfssl v1.4.0
	github.com/daaku/go.zipexe v1.0.1 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/jetstack/cert-manager v0.11.0
	github.com/lib/pq v1.2.0 // indirect
	github.com/mattn/go-sqlite3 v2.0.0+incompatible // indirect
	github.com/nkovacs/streamquote v1.0.0 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/valyala/fasttemplate v1.1.0 // indirect
	golang.org/x/tools v0.0.0-20191112005509-a3f652f18032 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apiextensions-apiserver v0.0.0-20190918201827-3de75813f604 // indirect
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/utils v0.0.0-20190801114015-581e00157fb1
	sigs.k8s.io/controller-runtime v0.3.1-0.20191022174215-ad57a976ffa1
)
