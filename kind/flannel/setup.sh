#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail
set -xv

if [[ ! -d plugins ]] ; then
    git clone https://github.com/containernetworking/plugins.git
    pushd plugins
      ./build_linux.sh
    popd
fi

CLUSTER_NAME=${CLUSTER_NAME:-netpol-flannel}
VERSION_FLANNEL="v0.13.0"

kind create cluster --name "$CLUSTER_NAME" --config conf.yaml
until kubectl cluster-info;  do
    echo "$(date)waiting for cluster..."
    sleep 2
done

kubectl get pods
kubectl apply -f https://raw.githubusercontent.com/flannel-io/flannel/v0.13.0/Documentation/kube-flannel.yml
sleep 5 ; kubectl -n kube-system get pods | grep flannel
echo "will wait for flannel to start running now... "
while true ; do
    kubectl -n kube-system get pods
    sleep 3
done
