#! /bin/bash

# Usage:
# ./.github/workflows/build.sh 

#first build should give permission to docker volumes
#sudo chmod -R 777 /var/lib/docker/volumes

set -e

UPDATE_BRANCH=${UPDATE_BRANCH:-"master"}
rm -rf istio-proxy
git clone -b ${UPDATE_BRANCH} https://github.com/intel/istio-proxy.git

pushd istio-proxy
# Update to release branch as intel/envoy has updated.
./scripts/update_envoy.sh
BUILD_WITH_CONTAINER=1 make build_envoy 
BUILD_WITH_CONTAINER=1 make exportcache
popd

# export env
TAG=${TAG:-"pre-build"}
make build
# replace upstream envoy with local envoy in build proxyv2 image
cp -rf istio-proxy/out/linux_amd64/envoy out/linux_amd64/release/envoy
# build proxyv2 image
make docker.proxyv2
