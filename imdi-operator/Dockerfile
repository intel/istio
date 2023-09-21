# Build the manager binary
FROM golang:1.20 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
RUN curl -LO https://dl.k8s.io/release/v1.27.4/bin/linux/amd64/kubectl
RUN curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.18.0 TARGET_ARCH=x86_64 sh -
RUN wget https://get.helm.sh/helm-v3.6.2-linux-amd64.tar.gz
RUN tar -zxvf helm-v3.6.2-linux-amd64.tar.gz
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/controller/ internal/controller/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
# RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager cmd/main.go

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager cmd/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM ubuntu:focal
WORKDIR /
RUN apt update && apt install -y git curl

COPY --from=builder /workspace/manager .
COPY --from=builder /workspace/kubectl /usr/local/bin/
COPY --from=builder /workspace/istio-*/bin/istioctl /usr/local/bin/
COPY --from=builder /workspace/linux-amd64/helm /usr/local/bin/
RUN chmod 755 /usr/local/bin/kubectl
RUN chmod 755 /usr/local/bin/istioctl
RUN chmod 755 /usr/local/bin/helm


ADD ./internal/chart internal/chart
ADD ./internal/istio/setup_istio.sh setup_istio.sh
RUN chmod 775 setup_istio.sh
ADD ./internal/istio/istio.yaml istio.yaml
ADD ./internal/istio/uninstall_istio.sh uninstall_istio.sh
RUN chmod 775 uninstall_istio.sh

ADD ./internal/deviceplugin/setup_deviceplugin_operator.sh setup_deviceplugin_operator.sh
RUN chmod 775 setup_deviceplugin_operator.sh
ADD ./internal/deviceplugin/setup_qat.sh setup_qat.sh
RUN chmod 775 setup_qat.sh
ADD ./internal/deviceplugin/qatdeviceplugin.yaml qatdeviceplugin.yaml
ADD ./internal/deviceplugin/setup_sgx_psw.sh setup_sgx_psw.sh
ADD ./internal/deviceplugin/setup_sgx.sh setup_sgx.sh
RUN chmod 775 setup_sgx.sh
RUN chmod 775 setup_sgx_psw.sh
ADD ./internal/deviceplugin/SgxDevicePlugin.yaml SgxDevicePlugin.yaml
ADD ./internal/deviceplugin/uninstall_deviceplugin_operator.sh uninstall_deviceplugin_operator.sh
RUN chmod 775 uninstall_deviceplugin_operator.sh
ADD ./internal/deviceplugin/uninstall_qat.sh uninstall_qat.sh
ADD ./internal/deviceplugin/uninstall_sgx.sh uninstall_sgx.sh
ADD ./internal/deviceplugin/uninstall_sgx_psw.sh uninstall_sgx_psw.sh
RUN chmod 775 uninstall_qat.sh
RUN chmod 775 uninstall_sgx.sh
RUN chmod 775 uninstall_sgx_psw.sh

ADD ./sgx-psw/sgx_aesmd sgx-psw/sgx_aesmd
RUN chmod 775 sgx-psw/sgx_aesmd

ADD ./internal/proxy/setup_proxy.sh setup_proxy.sh
ADD ./internal/proxy/proxy_test.sh proxy_test.sh
RUN chmod 775 setup_proxy.sh
RUN chmod 775 proxy_test.sh

# USER 65532:65532

Entrypoint ["/manager"]

