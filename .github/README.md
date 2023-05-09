# Intel managed distribution of Istio

## Introduction

Intel managed distribution of Istio is a project aiming to showcase and
integrate various Intel technologies into Istio and Envoy. The focus is
in letting both upstream community and users know what Intel is working
on, finding gaps in upstream project features in relation to hardware
enablement, and testing and deploying Intel features for Istio service
mesh.

Intel features are documented [here](https://intel.github.io/istio).

Intel managed distribution of Istio consists of the following source
code repositories:

* https://github.com/intel/envoy

* https://github.com/intel/envoy-go-control-plane

* https://github.com/intel/istio

* https://github.com/intel/istio-api

* https://github.com/intel/istio-proxy

## Project goals and relation to upstream projects

The goal of the project is not to maintain permanent Istio and Envoy
forks, but rather have a place to test and maintain features which can
be later upstreamed. When features are added to Istio and Envoy
upstream, they will be removed from this distribution, reducing the
difference to upstream. We intend to have Intel distribution for Istio
always compatible with the usual APIs, configuration formats, and
tooling. We will just extend Istio and Envoy with new extensions and
APIs.

## Release branch policy

Intel distribution for Istio will track the latest Istio release branch
in the corresponding branch. For example, Istio 1.15 will be tracked in
release-1.15-intel branch. The releases will be tagged with a naming
scheme such as 1.15.3-intel.2, which would indicate the second
Intel distribution for Istio release done on top of Istio upstream
1.15.3 release. The same Intel release branches and tags will be used
throughout other Intel projects that Istio depends on in the
repositories listed above. Keeping the branch and tag naming the same
over different repositories makes it clear which versions are used for
a particular Istio release.

## Features

### AVX-512

Envoy has support for CryptoMb private key provider plugin. This plugin
can be configured using Istio to accelerate TLS handshakes using RSA
for signing or decryption.

### Intel QuickAssist Technology (QAT) 2.0

QAT 2.0 (using QAT 4xxx generation devices present in future Xeon
Scalable processors) can be used to accelerate TLS handshakes using
RSA. QAT is also used to accelerate HTTP compression for gzip
encoding.

### SGX

SGX mTLS support helps maintain Envoy private key security by storing
the keys inside encrypted SGX enclaves.

## Deployment

TBD

## Development

This repository will be used in developing further support for
Intel technologies as briefly described above. The development
results will be made into Pull Requests for the upstream project.
Occationally upstream Pull Requests will be backported to earlier Istio
releases if needed. Ordinary Istio development shall take place in the
Istio upstream repository.

## Building

Istio with Intel enabled technologies is compiled from the top-level
source directory similar to an upstream build with:
```
TBD
```

## Upstream README

Upstream [README](/README.md).

## Limitations

This version of the software is pre-production release and is meant for
evaluation and trial purposes only.

## License

All of the source code required to build Intel managed distribution of Istio is available under Open Source licenses. The source code files identify the external Go modules that are used. Binaries are distributed as container images on DockerHub. Those images contain licenses text file `LICENSES.txt`. 