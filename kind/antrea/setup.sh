#!/usr/bin/env bash

set -xv
set -e

CLUSTER=${CLUSTER:-netpol-antrea}
VERSION=${VERSION:-v0.12.0}


if [[ ! -d antrea ]] ; then
  git clone https://github.com/vmware-tanzu/antrea.git
fi
pushd antrea
  git checkout "$VERSION"
  pushd ci/kind
    ./kind-setup.sh create "$CLUSTER"
  popd
popd
