# Envoy Nghttp2 HPACK acceleration

## Introduction

Nghttp2 implements HTTP/2 and HTTP/2 header compression algorithm HPACK. With the increasing size of HTTP headers in modern internet, more and more time are spent on HPACK (header compression). This solution we leverage AVX512 to accelerate Huffman Encoding in HPACK.

## Configuration

No extra configuration is needed. When the Nghttp2 patch is applied, the HPACK acceleration is enabled.

## Notification

Here is a list of notification during usage:

* Larger headers get more performance improvement.
* Header size shoud be less the 20 Kilo byte, otherwise it will fall back to original solution.
* Complex header (consist lot of non-alphabet characters) will get less performance improvement.
