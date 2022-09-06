# Istio mTLS Private Key Protection with SGX

## Introduction

Protecting Istio mTLS private keys with Intel® SGX enhances the service mesh security. The private keys are stored and used inside the SGX enclave(s) and will never stored in clear anywhere in the system. Authorized applications use the private key in the enclave by key-handle provided by SGX.

## Prerequisites

Prerequisites for using Istio mTLS private key protection with SGX:

- Kubernetes cluster with one or more nodes with Intel® [SGX](https://software.intel.com/content/www/us/en/develop/topics/software-guard-extensions.html) supported hardware
- [Intel® SGX device plugin](https://github.com/intel/intel-device-plugins-for-kubernetes/blob/main/cmd/sgx_plugin/README.md) for Kubernetes
- [Intel® SGX AESM daemon](https://github.com/intel/linux-sgx#install-the-intelr-sgx-psw)
- Linux kernel version 5.11 or later on the host (in tree SGX driver)

## Installation

This section covers how to install Istio mTLS private key protection with SGX

1. Install Istio

> NOTE: for the below command you need to use the `istioctl` for the `docker.io/intel/istioctl:1.15.1-intel.0` since only that contains Istio manifest enhancements for SGX mTLS.

You can also customize the `intel-istio-sgx-mTLS.yaml` according to your needs. If you want to enable sgx in sidecars or gateway, you can set the `sgx.enable` flag as `true`. And if you want do the quote verification, you should enable the `certExtensionValidationEnabled` flag.

```sh
istioctl install -f ./intel/yaml/intel-istio-sgx-mTLS.yaml -y
```

2. Verifiy the pods are running

By deault, `Istio` will be installed in the `istio-system` namespce

```sh
# Ensure that the pod is running state
$ kubectl get po -n istio-system
NAME                                    READY   STATUS    RESTARTS   AGE
istio-egressgateway-66b7c87ff8-qdf2z    1/1     Running   0          24s
istio-ingressgateway-789bb4b4f5-7mrq9   1/1     Running   0          24s
istiod-6d49576478-cg6xc                 1/1     Running   0          28s
```

## Deploy sample application

1. Create test namespace:

```sh
# create test namespace
$ kubectl create ns foo
```

2. Create httpbin deployment:

```sh
kubectl apply -f <(istioctl kube-inject -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml) -n foo
```

3. Create sleep deployment:

```sh
kubectl apply -f <(istioctl kube-inject -f https://raw.githubusercontent.com/istio/istio/master/samples/sleep/sleep.yaml) -n foo
```

4. Successful deployment looks like this:

```sh
$ kubectl get po -n foo
NAME                       READY   STATUS    RESTARTS   AGE
httpbin-7484c67b4d-8xf7q   2/2     Running     0        4m58s
sleep-6b74fd544d-q64j8     2/2     Running     0        5m13s
```
5. Test pod resources:

```sh
$ kubectl exec "$(kubectl get pod -l app=sleep -n foo -o jsonpath={.items..metadata.name})" -c sleep -n foo -- curl -s http://httpbin.foo:8000/headers | grep X-Forwarded-Client-Cert
    "X-Forwarded-Client-Cert": "By=spiffe://cluster.local/ns/foo/sa/httpbin;Hash=cd5d0504234e80c701c4fe01ef49f3fe048a63d1cdd5b9ffe3dd67ae3d93396b;Subject=\"CN=spiffe://cluster.local/ns/foo/sa/sleep\";URI=spiffe://cluster.local/ns/foo/sa/sleep"

```

The above `httpbin` and `sleep` applications have enabled SGX and store the private keys inside SGX enclave, completed the TLS handshake and established a connection with each other and communicating normally.

## Cleaning Up
```sh
# uninstall istio
$ istioctl x uninstall --purge -y
# delete workloads
$ kubectl delete ns foo
```

## See also

[Trusted Certificate Issuer](https://github.com/intel/trusted-certificate-issuer)

[Trusted Attestation Controller](https://github.com/intel/trusted-attestation-controller)
