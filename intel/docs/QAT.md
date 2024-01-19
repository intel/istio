# Istio crypto and compression acceleration with QAT

## Introduction

Intel® QuickAssist Technology (QAT) provides hardware acceleration for offloading security, authentication and compression services from the CPU, thus significantly increasing the performance and efficiency of standard platform solutions.

In this guide you will install Istio with the following QAT features enabled:

- QAT crypto acceleration for TLS handshakes
- QAT compression acceleration for HTTP(s) data

This solution is based on Linux in-tree driver and is utilizing the [qatlib](https://github.com/intel/qatlib)  and [qatzip](https://github.com/intel/qatzip) libraries.

## Acronyms

| Acronym | Description             |
|---------| ------------------------|
| QAT     | Intel® QuickAssist Technology available with 4th Gen Intel® Xeon® Scalable processors |
| cy      | Cryptographic |
| dc      | Compression |

## Prerequisites

Your Kubernetes nodes requires the following preparations

- Install Linux kernel 5.17 or similar
- Enable IOMMU from BIOS
- Enable IOMMU for Linux kernel
- Enhance the container runtime memory lock limit
- Install [Intel® QAT Device Plugin for Kubernetes](https://github.com/intel/intel-device-plugins-for-kubernetes)

To enable IOMMU for Linux kernel, add the following change and commands:

```console
cat /etc/default/grub:
GRUB_CMDLINE_LINUX="intel_iommu=on vfio-pci.ids=8086:4941"
update-grub
reboot
````

Once the system is rebooted, check if the IOMMU has been enabled via the following command:

```console
dmesg| grep IOMMU
[    1.528237] DMAR: IOMMU enabled
```

To enhance the `containerd` runtime memory lock limit, add the following file (CRIO has similar configuration):

```console
sudo mkdir /etc/systemd/system/containerd.service.d
sudo bash -c 'cat <<EOF >>/etc/systemd/system/containerd.service.d/memlock.conf
[Service]
LimitMEMLOCK=134217728
EOF'
```

Restart the container runtime (for containerd, CRIO has similar concept)

```console
sudo systemctl daemon-reload
sudo systemctl restart containerd
```

## Istio install with QAT

Clone the Intel managed distribution of Istio repo:

```
git clone -b 1.19.5-intel.0 --depth 1 https://github.com/intel/istio
```

Use the following command for the Istio installation:

```bash
istioctl install -y -f intel/yaml/intel-istio-qat-hw.yaml
```

The above command allocates single crypto (`qat.intel.com/cy`) and compression (`qat.intel.com/dc`) QAT endpoint for the `istio-ingress-gateway`. In addition, it defines Istio sidecar injection template (`sidecarInjectorWebhook`) for the sidecar QAT endpoint allocation.

At this stage, the `istio-ingress-gateway` is ready for QAT crypto acceleration for TLS handshakes.

To allocate QAT crypto endpoint for sidecars, add the following annotation to Kubernetes pods and/or deployments:

```console
inject.istio.io/templates: sidecar,qathw-crypto
```

With this annotation, the Istio sidecars are ready for QAT crypto acceleration for TLS handshakes.

To allocate QAT compression endpoint for sidecars, add the following annotation to Kubernetes pods and/or deployments:

```console
inject.istio.io/templates: sidecar,qathw-compression
```

Enable QAT compression acceleration for `istio-ingress-gateway`:

```console
kubectl apply -f intel/yaml/qat-compression-envoy-filter.yaml
```

At this stage, the `istio-ingress-gateway` is ready for QAT compression acceleration for HTTP(s) data.

Enable QAT compression acceleration for `istio-proxy` sidecars:

```console
kubectl apply -f  intel/yaml/compression-decompression-sidecar-envoy-filter.yaml
```

At this stage, the Istio sidecars are ready for QAT compression acceleration for HTTP(s) data.
