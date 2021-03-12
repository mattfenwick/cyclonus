#!/usr/bin/env bash

set -e
set -xv

CLUSTER=${CLUSTER:-netpol-cilium}
CILIUM_VERSION=${CILIUM_VERSION:-"v1.9.5"}
IMAGE=quay.io/cilium/cilium:${CILIUM_VERSION}

kind create cluster --name="$CLUSTER" --config=kind-config.yaml
until kubectl cluster-info; do
    echo "$(date) waiting for cluster..."
    sleep 2
done


helm repo add cilium https://helm.cilium.io/

docker pull "$IMAGE"
kind load docker-image "$IMAGE" --name "$CLUSTER"


helm install cilium cilium/cilium \
  --version "${CILIUM_VERSION}" \
  --namespace kube-system \
  --set nodeinit.enabled=true \
  --set kubeProxyReplacement=partial \
  --set hostServices.enabled=false \
  --set externalIPs.enabled=true \
  --set nodePort.enabled=true \
  --set hostPort.enabled=true \
  --set bpf.masquerade=false \
  --set image.pullPolicy=IfNotPresent \
  --set ipam.mode=kubernetes
