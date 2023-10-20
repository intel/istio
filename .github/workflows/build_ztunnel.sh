#! /bin/bash

# Usage:
# ./.github/workflows/build_ztunnel.sh 

#first build should give permission to docker volumes
#sudo chmod -R 777 /var/lib/docker/volumes

set -e

UPDATE_BRANCH=${UPDATE_BRANCH:-"master"}
rm -rf istio-ztunnel
git clone -b ${UPDATE_BRANCH} https://${USER}:${GITHUB_TOKEN}@github.com/intel-innersource/applications.services.cloud.istio.ztunnel.git istio-ztunnel

pushd istio-ztunnel
# Add  "--release" after cargo build in makefile
sed -i 's/cargo build/& --release/g' Makefile.core.mk
make build
popd

# export env
TAG=${TAG:-"pre-build"}
make build 
# replace upstream envoy with local envoy in build proxyv2 image
cp -rf istio-ztunnel/out/rust/release/ztunnel out/linux_amd64/release/ztunnel
# build proxyv2 image
VERSION=${TAG} make docker.ztunnel
