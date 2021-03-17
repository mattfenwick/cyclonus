#!/usr/bin/env bash

set -xv
set -eou pipefail

CLUSTER="netpol-ovn-kubernetes"
OVN_DIR="ovn-kubernetes-repo"

if [[ ! -d $OVN_DIR ]] ; then
  git clone https://github.com/ovn-org/ovn-kubernetes $OVN_DIR
fi

# TODO enable this or wait for https://github.com/ovn-org/ovn-kubernetes/pull/2112 to land
#cp patch-fedora33-cg0-enabled.patch ovn-kubernetes
pushd $OVN_DIR
#  patch -p1 < patch-fedora33-cg0-enabled.patch
  pushd go-controller
      make
  popd

  pushd dist/images
      make ubuntu
  popd

  pushd contrib
      KUBECONFIG=${HOME}/admin.conf ./kind.sh
  popd
popd
