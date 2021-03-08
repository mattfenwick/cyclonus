#!/usr/bin/env bash

set -e
set -xv

CLUSTER=${CLUSTER:-netpol-cilium}
CILIUM_VERSION="1.9.4"

kind create cluster --name="$CLUSTER" --config=conf.yaml
until kubectl cluster-info; do
    echo "$(date) waiting for cluster..."
    sleep 2
done


helm repo add cilium https://helm.cilium.io/

IMAGE=quay.io/cilium/cilium:v${CILIUM_VERSION}
docker pull $IMAGE
kind load docker-image $IMAGE --name "$CLUSTER"

# Install cilium with Helm
helm install cilium cilium/cilium \
  --version ${CILIUM_VERSION} \
  --namespace kube-system \
  --set nodeinit.enabled=true \
  --set kubeProxyReplacement=partial \
  --set hostServices.enabled=false \
  
# Fedora 33 Issue report
#-----------------------------
# coredns, local path storage containers stuck in CreatingContainer status forever with no
# endpoint available. 
#
# Workaround:
# 1) COMMENT LINE ABOVE (--set hostServices.enabled=false)
# 2) UNCOMMENT the line below
#  --set hostServices.enabled=true \
#  
# NOTE: The issue might affect others distros with kernel > 5.10
#
# Tested with:
#   - Fedora 33, Kernel: 5.10.19-200.fc33.x86_64
#
# See also: 
# https://github.com/cilium/cilium/issues/14960
# https://github.com/cilium/cilium/pull/14951
  
  --set externalIPs.enabled=true \
  --set nodePort.enabled=true \
  --set hostPort.enabled=true \
  --set bpf.masquerade=false \
  --set image.pullPolicy=IfNotPresent \
  --set ipam.mode=kubernetes
