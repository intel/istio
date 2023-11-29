# Envoy acceleration with QAT

## Introduction

Intel® QuickAssist Technology (QAT) provides hardware acceleration for offloading security, authentication and compression services from the CPU, thus significantly increasing the performance and efficiency of standard platform solutions.

In this guide you will learn how to Envoy with QAT crypto acceleration for TLS handshakes

This solution is based on Linux in-tree driver and is utilizing the [qatlib](https://github.com/intel/qatlib)

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

For running Envoy in the container, the `containerd` runtime memory lock limit need to be enhanced, add the following file (CRIO has similar configuration):

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

And the QAT vfio device (under `/dev/vfio`) should be passthrough to the container.

## Configuration

To enable QAT on HTTP1 or HTTP2, just as usual way to add [TLS Transportsocket](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/transport_sockets/tls/v3/tls.proto) for the downstream connection , but enable the [private key provider](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/transport_sockets/tls/v3/tls.proto). To enable QAT,
the [QAT provider](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/private_key_providers/qat/v3alpha/qat.proto) should be used. The configuration example for HTTP1/HTTP2 as below:

```yaml
  transport_socket:
    name: envoy.transport_sockets.tls
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.DownstreamTlsContext
      common_tls_context:
        tls_certificates:
        certificate_chain:
          filename: "/home/hejiexu/cert/server.pem"
        private_key_provider:
            provider_name: qat
            typed_config:  
              "@type": type.googleapis.com/envoy.extensions.private_key_providers.qat.v3alpha.QatPrivateKeyMethodConfig
              private_key:
                filename: "/home/hejiexu/cert/server-key.pem"
                poll_delay:
                  nanos: 5000000
```
