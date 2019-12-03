# CFSSL Issuer

[![Build Status](https://travis-ci.org/OpenSource-THG/cfssl-issuer.svg?branch=master)](https://travis-ci.org/OpenSource-THG/cfssl-issuer)
[![codecov](https://codecov.io/gh/OpenSource-THG/cfssl-issuer/branch/master/graph/badge.svg)](https://codecov.io/gh/OpenSource-THG/cfssl-issuer)
[![Go Report Card](https://goreportcard.com/badge/github.com/Opensource-THG/cfssl-issuer)](https://goreportcard.com/report/github.com/Opensource-THG/cfssl-issuer)
![Docker Pulls](https://img.shields.io/docker/pulls/opensourcethg/cfssl-issuer)

CFSSL Issuer is a controller that extends Jetstack's [cert-manager](https://github.com/jetstack/cert-manager) to add an issuer that uses a
CFSSL server to sign certificate requests.

## Installation

This controller requires a cert-manager version of > v0.11.0 and a running CFSSL server

### Helm

TBD

### Manually

```bash
git clone git@github.com:OpenSource-THG/cfssl-issuer.git
cd cfssl-issuer
kubectl apply -f deploy
```

## Configuration

Once installed we need to configure either a CfsslIssuer or CfsslClusterIssuer resource.

### Deployment

All CFSSL issuers share common configuraton for requesting certificates, namely the URL, Profile and CA Bundle

* URL is the url of a CFSSL server
* Profile is an optional field, denoting which profile cfssl should use when signing a Certificate
* CA Bundle is a base64 encoded string of the Certificate Authority to trust the CFSSL connection. The controller will
also asusme that this is the CA used when signing the Certificate Request

Below is an example of a namespaced and cluster scoped configuration

```yaml
kind: CfsslIssuer
apiVersion: certmanager.thg.io/v1beta1
metadata:
  name: cfsslissuer-server
spec:
  url: https://cfsslapi.local
  caBundle: <base64-encoded-ca>
```

```yaml
kind: CfsslClusterIssuer
apiVersion: certmanager.thg.io/v1beta1
metadata:
  name: cfsslissuer-server
spec:
  url: https://cfsslapi.local
  caBundle: <base64-encoded-ca>
```

The controller assumes that the cfssl api is secured via TLS using the provided CA Bundle and that the certs are signed by the same CA.

Certificates are then created via normal cert-manager flow referencing the issuer. As opposed to builtin issuers the group and kind
must be explicitly defined.

```yaml
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  name: example-com
spec:
  secretName: example-com-tls
  duration: 2160h # 90d
  renewBefore: 360h # 15d
  commonName: example.com
  dnsNames:
    - example.com
    - www.example.com
  issuerRef:
    name: cfsslissuer-server
    group: certmanager.thg.io
    kind: CfsslIssuer
```
