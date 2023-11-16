# Qat Zstd Compressor

Qat zstd compressor provides Envoy with faster hardware-accelerated zstd
compression by integrating with [Intel® QuickAssist Technology (Intel®
QAT)](https://www.intel.com/content/www/us/en/architecture-and-technology/intel-quick-assist-technology-overview.html)
through the qatlib and QAT-ZSTD-Plugin libraries.

## Example configuration

An example for Qat zstd compressor configuration is:

```
compressor_library:
  name: text_optimized
  typed_config:
    "@type": type.googleapis.com/envoy.extensions.compression.zstd.compressor.v3.Zstd
    compression_level: 10
    enable_qat_zstd: true
    qat_zstd_fallback_threshold: 0

```

## How it works

If enabled, the Qat zstd compressor will:

-   attach Qat hardware
-   create Threadlocal Qat zstd context for each worker thread

When a new http request comes, one worker thread will process it using its
Qat zstd context and send the data needed to be compressed to Qat
hardware using standard zstd api.

## Installing and using QAT-ZSTD-Plugin  

For information on how to build/install and use QAT-ZSTD-Plugin see
[introduction](https://github.com/intel/QAT-ZSTD-Plugin/tree/main#introduction).
