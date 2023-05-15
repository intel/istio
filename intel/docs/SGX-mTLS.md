# Istio mTLS Private Key Protection with SGX

## Introduction

Protecting Istio mTLS private keys with Intel® SGX enhances the service mesh security. The private keys are stored and used inside the SGX enclave(s) and will never stored in clear anywhere in the system. Authorized applications use the private key in the enclave by key-handle provided by SGX.

## Prerequisites

Prerequisites for using Istio mTLS private key protection with SGX:

- Kubernetes cluster with one or more nodes with [Intel® SGX](https://software.intel.com/content/www/us/en/develop/topics/software-guard-extensions.html) supported hardware
- [Intel® SGX device plugin for Kubernetes](https://github.com/intel/intel-device-plugins-for-kubernetes/blob/main/cmd/sgx_plugin/README.md)
- Linux kernel version 5.11 or later on the host (in tree SGX driver)
- [trusted-certificate-issuer](https://github.com/intel/trusted-certificate-issuer)
- [Intel® SGX AESM daemon](https://github.com/intel/linux-sgx#install-the-intelr-sgx-psw)
- [Intel® KMRA service](https://www.intel.com/content/www/us/en/developer/topic-technology/open/key-management-reference-application/overview.html)
- [Intel® Linux SGX](https://github.com/intel/linux-sgx) and [cripto-api-toolkit](https://github.com/intel/crypto-api-toolkit) in the host (optional, only needed if you want to build sds-server image locally)
> NOTE: The KMRA service and AESM daemon is also optional, needs to be set up only when remote attestaion required, which can be set through `NEED_QUOTE` flag in the chart.

## Installation

This section covers how to install Istio mTLS private key protection with SGX
1. Create sgx-signer (Refer to https://github.com/intel/trusted-certificate-issuer/blob/main/docs/istio-custom-ca-with-csr.md)

```sh
$ export CA_SIGNER_NAME=sgx-signer
$ cat << EOF | kubectl create -f -
apiVersion: tcs.intel.com/v1alpha1
kind: TCSClusterIssuer
metadata:
    name: $CA_SIGNER_NAME
spec:
    secretName: ${CA_SIGNER_NAME}-secret
    # If using quoteattestaion, set selfSign as false
    # selfSign: false
EOF
```

```sh
# Get CA Cert and replace it in ./intel/yaml/intel-istio-sgx-mTLS.yaml
$ kubectl get secret -n tcs-issuer ${CA_SIGNER_NAME}-secret -o jsonpath='{.data.tls\.crt}' |base64 -d | sed -e 's;\(.*\);        \1;g'
```
2. Build sds-server images

Build from source code: 
```sh
# Getting the source code
$ git clone https://github.com/istio-ecosystem/hsm-sds-server.git
```

```sh
# Build image
$ make docker
```
> NOTE: If you are using containerd as the container runtime, run `make ctr` to build the image instead.
3. Install Istio

> NOTE: for the below command you need to use the `istioctl` for the `docker.io/intel/istioctl:1.17.1-intel.2` since only that contains Istio manifest enhancements for SGX mTLS.

You can also customize the `intel-istio-sgx-mTLS.yaml` according to your needs. If you want do the quote verification, you can set the `NEED_QUOTE` env as `true`. And if you are using the TCS v1alpha1 api, you should set the `RANDOM_NONCE` as `false`.

```sh
$ istioctl install -f ./intel/yaml/intel-istio-sgx-mTLS.yaml -y
```

4. Verifiy the pods are running

By deault, `Istio` will be installed in the `istio-system` namespce

```sh
# Ensure that the pod is running state
$ kubectl get po -n istio-system
NAME                                    READY   STATUS    RESTARTS   AGE
istio-ingressgateway-6cd77bf4bf-t4cwj   1/1     Running   0          70m
istiod-6cf88b78dc-dthpw                 1/1     Running   0          70m
```

## Deploy sample application

1. Create sleep and httpbin deployment:
> NOTE: If you want use the sds-custom injection template, you need to set the annotations `inject.istio.io/templates` for both `sidecar` and `sgx`. And the ClusterRole is also required.
```sh
$ kubectl apply -f <(istioctl kube-inject -f ./intel/yaml/sleep-sgx-mTLS.yaml )
$ kubectl apply -f <(istioctl kube-inject -f ./intel/yaml/httpbin-sgx-mTLS.yaml )
```

1. Successful deployment looks like this:

```sh
$ kubectl get po
NAME                       READY   STATUS    RESTARTS   AGE
httpbin-5f6bf4d4d9-5jxj8   3/3     Running   0          30s
sleep-57bc8d74fc-2lw4n     3/3     Running   0          7s
```
3. Test pod resources:

```sh
$ kubectl exec "$(kubectl get pod -l app=sleep -o jsonpath={.items..metadata.name})" -c sleep -- curl -v -s http://httpbin.default:8000/headers | grep X-Forwarded-Client-Cert
    "X-Forwarded-Client-Cert": "By=spiffe://cluster.local/ns/default/sa/httpbin;Hash=2875ce095572f8a12b6080213f7789bfb699099b83e8ea2889a2d7b3eb9523e6;Subject=\"CN=SGX based workload,O=Intel(R) Corporation\";URI=spiffe://cluster.local/ns/default/sa/sleep"

```

The above `httpbin` and `sleep` applications have enabled SGX and store the private keys inside SGX enclave, completed the TLS handshake and established a connection with each other and communicating normally.

## Cleaning Up
```sh
# uninstall istio
$ istioctl x uninstall --purge -y
# delete workloads
$ kubectl delete -f ./intel/yaml/sleep-sgx-mTLS.yaml
$ kubectl delete -f ./intel/yaml/httpbin-sgx-mTLS.yaml
```

## See also

[Trusted Certificate Issuer](https://github.com/intel/trusted-certificate-issuer)

[Trusted Attestation Controller](https://github.com/intel/trusted-attestation-controller)