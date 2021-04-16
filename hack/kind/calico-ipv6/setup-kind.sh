#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail
set -xv

CLUSTER=${CLUSTER:-netpol-calico-ipv6}
PLATFORM=${PLATFORM:-"linux"}


if [[ $PLATFORM == "linux" ]]; then
  CONFIG=kind-config-linux.yaml
else
  CONFIG=kind-config-mac.yaml
fi

kind create cluster --name "$CLUSTER" --config $CONFIG


kubectl get nodes
kubectl get all -A

kubectl apply -f calico-3.18.1.yaml
# was: had to add 2 entries to calico configmap:
#   https://docs.projectcalico.org/networking/ipv6#enable-ipv6-only
#kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml

kubectl get nodes
kubectl get all -A

kubectl wait --for=condition=ready nodes --timeout=5m --all

kubectl get nodes
kubectl get all -A

kubectl wait --for=condition=ready pod -l k8s-app=calico-node -n kube-system
