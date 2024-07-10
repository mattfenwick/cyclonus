#!/usr/bin/env bash

set -xv
set -euo pipefail

KIND_VERSION=${KIND_VERSION:-v0.23.0}
CNI=${CNI:-calico}
CLUSTER_NAME="netpol-$CNI"
RUN_FROM_SOURCE=${RUN_FROM_SOURCE:-true}
FROM_SOURCE_ARGS=${FROM_SOURCE_ARGS:-"generate --include conflict --job-timeout-seconds 2"}
INSTALL_KIND=${INSTALL_KIND:-true}

# see https://github.com/actions/virtual-environments/blob/main/images/linux/Ubuntu2004-README.md
#   github includes a kind version, but it may not be the version we want
if [[ $INSTALL_KIND == true ]]; then
  curl -Lo ./kind https://github.com/kubernetes-sigs/kind/releases/download/"${KIND_VERSION}"/kind-$(uname)-amd64
  chmod +x ./kind
  sudo mv kind /usr/local/bin
fi

kind version
which -a kind

# create kind cluster
pushd "$CNI"
  CLUSTER=$CLUSTER_NAME ./setup-kind.sh
popd

# preload agnhost image
docker pull registry.k8s.io/e2e-test-images/agnhost:2.43
kind load docker-image registry.k8s.io/e2e-test-images/agnhost:2.43 --name "$CLUSTER_NAME"

# make sure that the new kind cluster is the current kubectl context
kind get clusters
kind export kubeconfig --name "$CLUSTER_NAME"

# get some debug info
kubectl get nodes
kubectl get pods -A

# run cyclonus
if [ "$RUN_FROM_SOURCE" == true ]; then
  # don't quote this -- we want word splitting here!
  go run ../../cmd/cyclonus/main.go $FROM_SOURCE_ARGS
else
  docker pull mfenwick100/cyclonus:latest
  kind load docker-image mfenwick100/cyclonus:latest --name "$CLUSTER_NAME"

  JOB_NAME=job.batch/cyclonus
  JOB_NS=netpol

  # set up cyclonus
  kubectl create ns "$JOB_NS"
  kubectl create clusterrolebinding cyclonus --clusterrole=cluster-admin --serviceaccount="$JOB_NS":cyclonus
  kubectl create sa cyclonus -n "$JOB_NS"

  pushd "$CNI"
    kubectl create -f cyclonus-job.yaml -n "$JOB_NS"
  popd

  # wait for job to start running
  # TODO there's got to be a better way to do this
  sleep 30
  kubectl get all -A

  kubectl wait --for=condition=ready pod -l job-name=cyclonus -n $JOB_NS --timeout=5m

  kubectl logs -f -n $JOB_NS $JOB_NAME
fi
