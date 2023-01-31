# Istio Gateway Private Key Protection with SGX

## Introduction

Protecting Istio gateway private keys with Intel® SGX enhances the service mesh security. The private keys are stored and used inside the SGX enclave(s) and will never stored in clear anywhere in the system. Authorized applications use the private key in the enclave by key-handle provided by SGX.

## Prerequisites

Prerequisites for using Istio gateway private key protection with SGX:

- Kubernetes cluster with one or more nodes with Intel® [SGX](https://software.intel.com/content/www/us/en/develop/topics/software-guard-extensions.html) supported hardware
- [Intel® SGX device plugin](https://github.com/intel/intel-device-plugins-for-kubernetes/blob/main/cmd/sgx_plugin/README.md) for Kubernetes
- [Intel® SGX AESM daemon](https://github.com/intel/linux-sgx#install-the-intelr-sgx-psw)
- Linux kernel version 5.11 or later on the host (in tree SGX driver)
- [trusted-certificate-issuer](https://github.com/intel/trusted-certificate-issuer)

## Installation

This section covers how to install Istio gateway private key protection with SGX
1. Install Istio

> NOTE: for the below command you need to use the `istioctl` for the `docker.io/intel/istioctl:1.16.1-intel.0` since only that contains Istio manifest enhancements for SGX mTLS.
You can also customize the `intel-istio-sgx-gateway.yaml` according to your needs. If you want do the quote verification, you can set the `NEED_QUOTE` env as `true`. And if you are using the TCS v1alpha1 api, you should set the `RANDOM_NONCE` as `false`.

```sh
istioctl install -f ./intel/yaml/intel-istio-sgx-gateway.yaml -y
```

2. Verifiy the pods are running

By deault, `Istio` will be installed in the `istio-system` namespce

```sh
# Ensure that the pods are running state
$ kubectl get pod -n istio-system
NAME                                    READY   STATUS    RESTARTS   AGE
istio-ingressgateway-55f8dbb66c-6qx2s   2/2     Running   0          73s
istiod-65db6d8666-jgmf7                 1/1     Running   0          75s
```

## Deploy sample application

1. Create httpbin deployment with gateway CR:
> NOTE: If you want use the sds-custom injection template, you need to set the annotations `inject.istio.io/templates` for both `sidecar` and `sgx`. And the ClusterRole is also required.
```sh
kubectl apply -f <(istioctl kube-inject -f ./intel/yaml/httpbin-sgx-gateway.yaml )
kubectl apply -f <(istioctl kube-inject -f ./intel/yaml/httpbinGW-sgx-gateway.yaml )
```
2. Successful deployment looks like this:

Verify the httpbin pod:
```sh
$ kubectl get pod -n default
NAME                       READY   STATUS    RESTARTS      AGE
httpbin-7fbf9db8f6-qvqn4   3/3     Running   4 (97s ago)   2m27s
```

Verify the gateway CR:
```sh
$ kubectl get gateway -n default
NAME              AGE
testuds-gateway   2m52s
```

Verify the quoteattestation CR:
```sh
$ kubectl get quoteattestations.tcs.intel.com -n default
NAME                                                                            AGE
sgxquoteattestation-istio-ingressgateway-55f8dbb66c-6qx2s-httpbin-testsds-com   4m36s
```
Manually get the quoteattestation name via below command

```sh
$ export QA_NAME=<YOUR QUOTEATTESTATION NAME>
```
3. Prepare credential information:

We use command line tools to read and write the QuoteAttestation manually. You get the tools, `km-attest` and `km-wrap`, provided by the [Intel® KMRA project](https://www.intel.com/content/www/us/en/developer/topic-technology/open/key-management-reference-application/overview.html).

> NOTE: please use release version 2.2.1

```sh
$ mkdir -p $HOME/sgx/gateway
$ export CREDENTIAL=$HOME/sgx/gateway

$ kubectl get quoteattestations.tcs.intel.com -n default $QA_NAME -o jsonpath='{.spec.publicKey}' | base64 -d > $CREDENTIAL/public.key
$ kubectl get quoteattestations.tcs.intel.com -n default $QA_NAME -o jsonpath='{.spec.quote}' | base64 -d > $CREDENTIAL/quote.data
$ km-attest --pubkey $CREDENTIAL/public.key --quote $CREDENTIAL/quote.data

$ openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout $CREDENTIAL/example.com.key -out $CREDENTIAL/example.com.crt
$ openssl req -out $CREDENTIAL/httpbin.csr -newkey rsa:2048 -nodes -keyout $CREDENTIAL/httpbin.key -subj "/CN=httpbin.example.com/O=httpbin organization"
$ openssl x509 -req -sha256 -days 365 -CA $CREDENTIAL/example.com.crt -CAkey $CREDENTIAL/example.com.key -set_serial 0 -in $CREDENTIAL/httpbin.csr -out $CREDENTIAL/httpbin.crt
```
> NOTE: Before using `km-attest`, please configurate `/opt/intel/km-wrap/km-wrap.conf` according to below content:
```
{
    "keys": [
        {
            "signer": "tcsclusterissuer.tcs.intel.com/sgx-signer",
            "key_path": "$CREDENTIAL/httpbin.key",
            "cert": "$CREDENTIAL/httpbin.crt"
        }
    ]
}
```

4. Update credential quote attestation CR with secret contained wrapped key

```sh
$ WRAPPED_KEY=$(km-wrap --signer tcsclusterissuer.tcs.intel.com/sgx-signer --pubkey $CREDENTIAL/public.key --pin "HSMUserPin" --token "HSMSDSServer" --module /usr/local/lib/softhsm/libsofthsm2.so)

$ kubectl create secret generic -n default wrapped-key --from-literal=tls.key=${WRAPPED_KEY} --from-literal=tls.crt=$(base64 -w 0 < $CREDENTIAL/httpbin.crt)
```
Edit quoteattestations.tcs.intel.com $QA_NAME via commond `kubectl edit quoteattestations.tcs.intel.com $QA_NAME -n default` and append field `secretName: wrapped-key` for it.

The above `httpbin` applications have enabled SGX and store the private keys inside SGX enclave, completed the TLS handshake and established a connection with each other and communicating normally.

5. Verify the service accessibility

```sh
$ export INGRESS_NAME=istio-ingressgateway
$ export INGRESS_NS=istio-system
$ export SECURE_INGRESS_PORT=$(kubectl -n "${INGRESS_NS}" get service "${INGRESS_NAME}" -o jsonpath='{.spec.ports[?(@.name=="https")].nodePort}')
$ export INGRESS_HOST=$(kubectl get po -l istio=ingressgateway -n "${INGRESS_NS}" -o jsonpath='{.items[0].status.hostIP}')

$ curl -v -HHost:httpbin.example.com --resolve "httpbin.example.com:$SECURE_INGRESS_PORT:$INGRESS_HOST" \
  --cacert $CREDENTIAL/example.com.crt "https://httpbin.example.com:$SECURE_INGRESS_PORT/status/418"
```
It will be okay if got below response:
[Response](../image/gateway-test.png)

## Cleaning Up
```sh
# delete workloads and secret
$ kubectl delete -f ./intel/yaml/httpbin-sgx-gateway.yaml -n default
$ kubectl delete -f ./intel/yaml/httpbinGW-sgx-gateway.yaml -n default
# uninstall istio
$ istioctl x uninstall --purge -y
```

## See also

[Trusted Certificate Issuer](https://github.com/intel/trusted-certificate-issuer)

[Trusted Attestation Controller](https://github.com/intel/trusted-attestation-controller)
