# Envoy acceleration with Intel® AVX512 crypto instructions

## Introduction

Cryptographic operations are among the most compute-intensive and critical operations when it comes to secured connections. The TLS termination is a key feature of Envoy,
and special in the use-case of Envoy using as a gateway. The TLS handshake becomes the most compute-intensive and critical operations.

CryptoMB means using Intel® Advanced Vector Extensions 512 (Intel® AVX-512) instructions using a SIMD (single instruction, multiple data) mechanism. Up to eight RSA or ECDSA operations are gathered into a buffer and processed at the same time, providing potentially improved performance. Intel AVX-512 instructions are available on recently launched 3rd generation Intel Xeon Scalable processor server processors, or later.

In this document you will learn how to enable CryptoMB in Envoy to achieve performance improvements and accelerated handshakes.

## Prerequisites

- At least one node 3rd generation Intel Xeon Scalable processor server processors, or later.  
  And the following instructions are required to use CryptoMB:
  - BMI2
  - AVX512F
  - AVX512DQ
  - AVX512BW
  - AVX512IFMA
  - AVX512VBMI2
  - AVX512_ENABLEDBYOS

## Configuration

To enable CryptoMB on HTTP1 or HTTP2, just as usual way to add [TLS Transportsocket](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/transport_sockets/tls/v3/tls.proto) for the downstream connection , but enable the [private key provider](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/transport_sockets/tls/v3/tls.proto). To enable CryptoMB,
the [CryptoMB provider](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/contrib/cryptomb/cryptomb) should be used. The configuration example for HTTP1 and HTTP2 as below:

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
          provider_name: cryptomb
          typed_config:  
            "@type": type.googleapis.com/envoy.extensions.private_key_providers.cryptomb.v3alpha.CryptoMbPrivateKeyMethodConfig 
            private_key:
              filename: "/home/hejiexu/cert/server-key.pem"
            poll_delay:
              nanos: 5000000
```

To enable CryptoMB on HTTP3(QUIC), the [QUIC downstream transport socket](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/transport_sockets/quic/v3/quic_transport.proto#extensions-transport-sockets-quic-v3-quicdownstreamtransport) should be used, all other parts are same with HTTP1 and HTTP2. The configuration example for HTTP3(QUIC) as below:
```yaml
  transport_socket:
    name: envoy.transport_sockets.tls
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.transport_sockets.quic.v3.QuicDownstreamTransport
      common_tls_context:
        tls_certificates:
        certificate_chain:
          filename: "/home/hejiexu/cert/server.pem"
        private_key_provider:
          provider_name: cryptomb
          typed_config:  
            "@type": type.googleapis.com/envoy.extensions.private_key_providers.cryptomb.v3alpha.CryptoMbPrivateKeyMethodConfig 
            private_key:
              filename: "/home/hejiexu/cert/server-key.pem"
            poll_delay:
              nanos: 5000000
```