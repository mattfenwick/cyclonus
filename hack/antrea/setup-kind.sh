#!/usr/bin/env bash

set -xv
set -e

CLUSTER=${CLUSTER:-netpol-antrea}
VERSION=${VERSION:-v0.12.2}
ANTREA_DIR=antrea-repo

if [[ ! -d $ANTREA_DIR ]] ; then
  git clone https://github.com/vmware-tanzu/antrea.git $ANTREA_DIR
fi
pushd $ANTREA_DIR
  git checkout "$VERSION"
  pushd ci/kind
    ./kind-setup.sh create "$CLUSTER"
  popd
popd
