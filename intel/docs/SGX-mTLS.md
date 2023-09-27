# Istio mTLS Private Key Protection with SGX

## Introduction

Protecting Istio mTLS private keys with Intel® SGX enhances the service mesh security. The private keys are stored and used inside the SGX enclave(s) and will never stored in clear anywhere in the system. Authorized applications use the private key in the enclave by key-handle provided by SGX. For more application scenarios, please refer to [this document](https://github.com/istio-ecosystem/hsm-sds-server/blob/main/README.md)

## Prerequisites

Prerequisites for using Istio mTLS private key protection with SGX:

- [Intel® SGX software stack](setup-sgx-software.md)
- Kubernetes cluster with one or more nodes with [Intel® SGX](https://software.intel.com/content/www/us/en/develop/topics/software-guard-extensions.html) supported hardware
- [Intel® SGX device plugin for Kubernetes](https://github.com/intel/intel-device-plugins-for-kubernetes/blob/main/cmd/sgx_plugin/README.md)
- Linux kernel version 5.11 or later on the host (in tree SGX driver)
- Custom CA which support [Kubernetes CSR API](https://kubernetes.io/docs/reference/access-authn-authz/certificate-signing-requests/)
- [Intel® KMRA service](https://www.intel.com/content/www/us/en/developer/topic-technology/open/key-management-reference-application/overview.html) (Optional, needs to be set up only when remote attestation required, which can be set through `NEED_QUOTE` flag in the chart.)
- [cripto-api-toolkit](https://github.com/intel/crypto-api-toolkit) in the host (optional, only needed if you want to build sds-server image locally)
> NOTE: The KMRA service and AESM daemon is also optional, needs to be set up only when remote attestaion required, which can be set through `NEED_QUOTE` flag in the chart.

## Installation

This section covers how to install Istio mTLS private key protection with SGX. We use Cert Manager as default K8s CA in this document. If you want to use TCS for remote attestaion, please refer to this [Document](https://github.com/istio-ecosystem/hsm-sds-server/blob/main/Install-with-TCS.md).

> Note: please ensure installed cert manager with flag  `--feature-gates=ExperimentalCertificateSigningRequestControllers=true`. You can use `--set featureGates="ExperimentalCertificateSigningRequestControllers=true"` when helm install cert-manager

### Create signer

```sh
$ cat <<EOF > ./istio-cm-issuer.yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-istio-issuer
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: istio-ca
  namespace: cert-manager
spec:
  isCA: true
  commonName: istio-system
  secretName: istio-ca-selfsigned
  issuerRef:
    name: selfsigned-istio-issuer
    kind: ClusterIssuer
    group: cert-manager.io
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: istio-system
spec:
  ca:
    secretName: istio-ca-selfsigned
EOF
$ kubectl apply -f ./istio-cm-issuer.yaml
```

```sh
# Get CA Cert and replace it in ./deployment/istio-configs/istio-hsm-config.yaml
$ kubectl get clusterissuers istio-system -o jsonpath='{.spec.ca.secretName}' | xargs kubectl get secret -n cert-manager -o jsonpath='{.data.ca\.crt}' | base64 -d
```

### Apply quote attestation CRD
```sh
$ kubectl apply -f https://github.com/intel/trusted-certificate-issuer/tree/main/deployment/crds
```

### Protect the private keys of workloads with HSM
- Install Istio

```sh
$ istioctl install -f ./deployment/istio-configs/istio-hsm-config.yaml -y
```

- Verifiy the Istio is ready

By deault, `Istio` will be installed in the `istio-system` namespce

```sh
# Ensure that the pod is running state
$ kubectl get po -n istio-system
NAME                                    READY   STATUS    RESTARTS   AGE
istio-ingressgateway-6cd77bf4bf-t4cwj   1/1     Running   0          70m
istiod-6cf88b78dc-dthpw                 1/1     Running   0          70m
```

- Create sleep and httpbin deployment:
> NOTE: If you want use the sds-custom injection template, you need to set the annotations `inject.istio.io/templates` for both `sidecar` and `sgx`. And the ClusterRole is also required.
```sh
$ kubectl apply -f <(istioctl kube-inject -f ./deployment/istio-configs/sleep-hsm.yaml )
$ kubectl apply -f <(istioctl kube-inject -f ./deployment/istio-configs/httpbin-hsm.yaml )
```

> A reminder, if you want to apply other workloads, please make sure to add the correct RBAC rules for its `Service Account`. For details, please refer to the configuration of `ClusterRole` in `./deployment/istio-configs/httpbin-hsm.yaml`.

- Successful deployment looks like this:

```sh
$ kubectl get po
NAME                       READY   STATUS    RESTARTS   AGE
httpbin-5f6bf4d4d9-5jxj8   3/3     Running   0          30s
sleep-57bc8d74fc-2lw4n     3/3     Running   0          7s
```
- Test pod resources:

```sh
$ kubectl exec "$(kubectl get pod -l app=sleep -o jsonpath={.items..metadata.name})" -c sleep -- curl -v -s http://httpbin.default:8000/headers | grep X-Forwarded-Client-Cert
    "X-Forwarded-Client-Cert": "By=spiffe://cluster.local/ns/default/sa/httpbin;Hash=2875ce095572f8a12b6080213f7789bfb699099b83e8ea2889a2d7b3eb9523e6;Subject=\"CN=SGX based workload,O=Intel(R) Corporation\";URI=spiffe://cluster.local/ns/default/sa/sleep"

```

The above `httpbin` and `sleep` applications have enabled SGX and store the private keys inside SGX enclave, completed the TLS handshake and established a connection with each other.

```sh
# Dump the envoy config
$ kubectl exec "$(kubectl get pod -l app=sleep -o jsonpath={.items..metadata.name})" -c istio-proxy -- bash 

$ curl localhost:15000/config_dump > envoy_conf.json
```
It can be seen from the config file that the `private_key_provider` configuation has replaced the original private key, and the real private key has been safely stored in the SGX enclave.

## See also

[Trusted Certificate Issuer](https://github.com/intel/trusted-certificate-issuer)

[Trusted Attestation Controller](https://github.com/intel/trusted-attestation-controller)
