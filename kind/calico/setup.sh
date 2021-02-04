#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail
set -xv

CLUSTER_NAME=${CLUSTER_NAME:-netpol-calico}


kind create cluster --name "$CLUSTER_NAME" --config conf.yaml
until kubectl cluster-info;  do
    echo "$(date)waiting for cluster..."
    sleep 2
done


kubectl get pods
kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml
kubectl -n kube-system set env daemonset/calico-node FELIX_IGNORELOOSERPF=true
kubectl -n kube-system set env daemonset/calico-node FELIX_XDPENABLED=false
sleep 5 ; kubectl -n kube-system get pods | grep calico-node
echo "will wait for calico to start running now... "
while true ; do
    kubectl -n kube-system get pods
    sleep 3
done
