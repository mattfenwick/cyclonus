#!/usr/bin/env bash

set -e

CLUSTER=${CLUSTER:-netpol-cilium}
CONFIG=${CONFIG:-netpol-cilium-conf.yaml}

cat << EOF > "$CONFIG"
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
- role: worker
networking:
  disableDefaultCNI: true
EOF



kind create cluster --name="$CLUSTER" --config="$CONFIG"
until kubectl cluster-info;  do
    echo "$(date) waiting for cluster..."
    sleep 2
done


CILIUM_VERSION="1.9.1"

# Add Cilium Helm repo
helm repo add cilium https://helm.cilium.io/

# Pre-load images
docker pull cilium/cilium:"v${CILIUM_VERSION}"

# Install cilium with Helm
helm install cilium cilium/cilium \
  --version ${CILIUM_VERSION} \
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
