# Istio acceleration with Intel® AVX512 crypto instructions

## Introduction

Cryptographic operations are among the most compute-intensive and critical operations when it comes to secured connections. Istio uses Envoy as the “gateways/sidecar” to handle secure connections and intercept the traffic.

Depending upon use cases, when an ingress gateway must handle a large number of incoming TLS and secured service-to-service connections through sidecar proxies, the load on Envoy increases. The potential performance depends on many factors, such as size of the cpuset on which Envoy is running, incoming traffic patterns, and key size. These factors can impact Envoy serving many new incoming TLS requests. In this document you will learn how to enable CryptoMB in Istio to achieve performance improvements and accelerated handshakes.

CryptoMB means using Intel® Advanced Vector Extensions 512 (Intel® AVX-512) instructions using a SIMD (single instruction, multiple data) mechanism. Up to eight RSA or ECDSA operations are gathered into a buffer and processed at the same time, providing potentially improved performance. Intel AVX-512 instructions are available on recently launched 3rd generation Intel Xeon Scalable processor server processors, or later.

## Prerequisites

- Kubernetes cluster with at least one node 3rd generation Intel Xeon Scalable processor server processors, or later.  
  If not all nodes that support Intel® AVX-512 in Kubernetes cluster, you need to add some labels to divide these two kinds of nodes manually or using [NFD](https://github.com/kubernetes-sigs/node-feature-discovery).  
  And the following instructions are required to use CryptoMB:
  - BMI2
  - AVX512F
  - AVX512DQ
  - AVX512BW
  - AVX512IFMA
  - AVX512VBMI2
  - AVX512_ENABLEDBYOS

- Istio 1.14, or later

## Install

Clone the Intel managed distribution of Istio repo:

```
git clone -b 1.19.5-intel.0 --depth 1 https://github.com/intel/istio
```

Use the following command for the installation:

```bash
istioctl install -y -f intel/yaml/intel-istio-cryptomb.yaml
```

With the above installation, CryptoMB acceleration is used both in `istio-ingress-gateway` and `istio-proxy` sidecar containers.

## See also

[CryptoMB - TLS handshake acceleration for Istio](https://istio.io/latest/blog/2022/cryptomb-privatekeyprovider)

[A snapshot of Intel’s contributions to Istio](https://www.intel.com/content/www/us/en/developer/articles/technical/a-snapshot-of-contributions-to-istio.html)
