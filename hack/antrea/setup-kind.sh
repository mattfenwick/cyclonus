#!/usr/bin/env bash

set -xv
set -e

CLUSTER=${CLUSTER:-netpol-antrea}
VERSION=${VERSION:-v0.13.1}
ANTREA_DIR=antrea-repo

if [[ ! -d $ANTREA_DIR ]] ; then
  git clone https://github.com/vmware-tanzu/antrea.git $ANTREA_DIR
fi
pushd $ANTREA_DIR
  git checkout "$VERSION"

  pushd ci/kind
    ./kind-setup.sh create "$CLUSTER" --antrea-cni false
  popd

  pushd hack
    ./generate-manifest.sh --kind --tun vxlan | kubectl apply --context "kind-${CLUSTER}" -f -
  popd
popd
