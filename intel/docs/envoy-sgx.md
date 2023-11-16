# Private key protection with Intel SGX

## Introduction

One of benefit introduced by Istio/Envoy is data traffic encryption. However, the private keys which are the most critical assets in data security are not well protected, they either in pure text in file system or system memory without any encryption. To resolve this problem, we leverage Intel SGX ([Intel® Software Guard Extensions (Intel® SGX)](https://www.intel.com/content/www/us/en/architecture-and-technology/software-guard-extensions.html)) and put private keys into SGX enclave in which the memory is encrypted to provide protection.

## Configuration

The solution is preferred to use with Istio (not recommended to use standalone with Envoy due to the complicated configuration), please check this [link ](https://github.com/istio-ecosystem/hsm-sds-server/blob/main/README.md)for reference.
