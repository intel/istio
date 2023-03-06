##The new features are
Excluding Istio interfaces for 5G core
Istio mTLS key protection with SGX (preview of new the architecture)
Intel Istio Grafana dashboard integration with Meshery (preview)
Envoy TLS Bumping and splicing (preview)
Trusted Attestation Controller support for Intel Security Libraries (iSecL)
Helm chart for Trusted Attestation Controller
Envoy zSTD QAT acceleration (preview)

##The updated features are
Istio TLS handshake acceleration with ICX AVX512 crypto
Istio crypto and compression acceleration with QAT2.0
Trusted certificate issuer
Trusted attestation controller
Istio multi-tenancy: Multiple CAs
Istio modsecurity WASM plugin
Bypass TCP/IP stack using eBPF
Grafana dashboard for CryptoMB

##Supported versions
Kubernetes v1.24
Istio v1.16.0
cert-manager v1.7 or later
Linux kernel 5.11 or later (5.15 for QAT compression)
Intel Device Plugins for Kubernetes v0.25.1

##Install
Intel Istio can be installed with the following command:

istioctl install -y --set hub=docker.io/intel --set tag=1.16.0-intel.0