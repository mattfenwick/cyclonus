#!/usr/bin/env bash

set -xv
set -e

CLUSTER=${CLUSTER:-netpol-antrea}
VERSION=${VERSION:-v1.7.0}
ANTREA_DIR=antrea-repo
IMG_NAME="projects.registry.vmware.com/antrea/antrea-ubuntu"

if [[ ! -d $ANTREA_DIR ]] ; then
  git clone https://github.com/vmware-tanzu/antrea.git $ANTREA_DIR
fi
pushd $ANTREA_DIR
  git checkout "$VERSION"

  pushd ci/kind
    ./kind-setup.sh create "$CLUSTER" --antrea-cni false
  popd

  pushd hack
    IMG_NAME="${IMG_NAME}" IMG_TAG="${VERSION}" ./generate-manifest.sh --mode release --kind --tun vxlan | kubectl apply --context "kind-${CLUSTER}" -f -
  popd
popd
