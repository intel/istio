#! /bin/bash

# Usage:
# ./.github/workflows/build.sh 

# this script will modified upper directory istio-proxy repo
# if self-used should attention that whether the upper directory have its own developing istio.proxy
#first build should give permission to docker volumes
#sudo chmod -R 777 /var/lib/docker/volumes

UPDATE_BRANCH=${UPDATE_BRANCH:-"master"}
cd ..
rm -rf istio-proxy
git clone -b ${UPDATE_BRANCH} https://github.com/intel/istio-proxy.git
cp -rf istio/ istio-proxy/ 
cd istio-proxy

# Update to release branch as intel/envoy has updated.
./scripts/update_envoy.sh
BUILD_WITH_CONTAINER=1 make build_envoy 
BUILD_WITH_CONTAINER=1 make exportcache

cd istio
# export env
TAG=${TAG:-"pre-build"}
make build
# replace upstream envoy with local envoy in build proxyv2 image
cd  ..
cp -rf out/linux_amd64/envoy istio/out/linux_amd64/release/envoy
# build proxyv2 image
cd istio
make docker.proxyv2
