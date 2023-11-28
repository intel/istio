# Intel managed distribution of Envoy

Intel managed distribution of Envoy is a project aiming to showcase and integrate various Intel technologies into Envoy. The focus is in letting both upstream community and users know what Intel is working on, finding gaps in upstream project features in relation to hardware enablement, and testing and deploying Intel features for Envoy.

## Features

* [TLS handshake acceleration with AVX512 (CryptoMB)](docs/envoy-cryptomb.md)
* [TLS handshake acceleration with QAT2.0](docs/envoy-qat.md)
* [HTTP data compression acceleration with QATzip](docs/envoy-qatzip.md)
* [HTTP data compression acceleration with QATzstd](docs/envoy-qatzstd.md)
* [Regular expressions with Hyperscan](docs/envoy-hyperscan.md)
* [HTTP/2 and HTTP/2 header compression algorithm HPACK with AVX512](docs/envoy-nghttp2.md)
* [Private key protection with SGX](docs/envoy-sgx.md)
* [Connection balancer with DLB](https://www.envoyproxy.io/docs/envoy/latest/configuration/other_features/dlb)

## build

Use the following command for building envoy with Intel features:

```bash
# run test and build
./ci/run_envoy_docker.sh "./ci/do_ci.sh release"  

# only build envoy
./ci/run_envoy_docker.sh "./ci/do_ci.sh release.server_only" 
```

# Intel managed distribution of Istio

Intel managed distribution of Istio is a project aiming to showcase and integrate various Intel technologies into Istio with Intel managed distribution of Envoy. The focus is in letting both upstream community and users know what Intel is working on, finding gaps in upstream project features in relation to hardware enablement, and testing and deploying Intel features for Istio service mesh.

## Features

* [TLS handshake acceleration with AVX512 (CryptoMB)](docs/CRYPTOMB.md)
* [TLS handshake acceleration with QAT2.0](docs/QAT.md)
* [HTTP data compression acceleration with QAT2.0](docs/QAT.md)
* [mTLS Private key protection with SGX](docs/SGX-mTLS.md)
* [Gateway Private key protection with SGX](docs/SGX-gateway.md)
* [Grafana Dashboard](docs/Grafana-Dashboard.md)
* [Excluding Istio interfaces for 5G core](docs/Excluding-Istio-interfaces-for-5G-core.md)
* [Use Meshery to deploy and manage Istio with Intel Features](docs/Meshery.md)
* [Kiali](docs/Kiali.md)

## Supported versions
* Kubernetes v1.24
* Istio v1.15
* cert-manager v1.7 or later
* Linux kernel 5.18 or later
* Intel Device Plugins for Kubernetes v0.24

## Install

Use the follwoing command for basic installation:

```bash
istioctl install --set hub=docker.io/intel --set tag=1.15.1-intel.0
