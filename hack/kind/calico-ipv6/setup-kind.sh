#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail
set -xv

CLUSTER=${CLUSTER:-netpol-calico-ipv6}


kind create cluster --name "$CLUSTER" --config kind-config.yaml
until kubectl cluster-info;  do
    echo "$(date)waiting for cluster..."
    sleep 2
done


kubectl get nodes
kubectl get all -A

kubectl apply -f calico.yaml
# was: had to add 2 entries to calico configmap:
#   https://docs.projectcalico.org/networking/ipv6#enable-ipv6-only
#kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml

kubectl -n kube-system set env daemonset/calico-node FELIX_XDPENABLED=false
kubectl -n kube-system set env daemonset/calico-node FELIX_IPV6SUPPORT=true
kubectl -n kube-system set env daemonset/calico-node IP6=autodetect

kubectl get nodes
kubectl get all -A

kubectl wait --for=condition=ready nodes --timeout=5m --all

kubectl get nodes
kubectl get all -A

kubectl wait --for=condition=ready pod -l k8s-app=calico-node -n kube-system
