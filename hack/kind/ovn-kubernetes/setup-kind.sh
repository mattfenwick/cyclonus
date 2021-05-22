#!/usr/bin/env bash

set -xv
set -euo pipefail

CLUSTER=${CLUSTER:-"ovn"}
OVN_DIR="ovn-kubernetes-repo"

if [[ ! -d $OVN_DIR ]] ; then
  git clone https://github.com/ovn-org/ovn-kubernetes $OVN_DIR
fi

pushd $OVN_DIR
  pushd go-controller
      make
  popd

  pushd dist/images
      make ubuntu
  popd

  pushd contrib
      KIND_CLUSTER_NAME=$CLUSTER KUBECONFIG=${HOME}/admin.conf ./kind.sh
  popd
popd
