#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail
set -xv

CLUSTER_NAME=${CLUSTER_NAME:-netpol-calico}


kind create cluster --name "$CLUSTER_NAME" --config kind-config.yaml
until kubectl cluster-info;  do
    echo "$(date)waiting for cluster..."
    sleep 2
done


kubectl get pods
kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml
kubectl -n kube-system set env daemonset/calico-node FELIX_IGNORELOOSERPF=true
kubectl -n kube-system set env daemonset/calico-node FELIX_XDPENABLED=false


kubectl wait --for=condition=ready pod -l k8s-app=calico-node -n kube-system
