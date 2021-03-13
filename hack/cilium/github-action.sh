#!/usr/bin/env bash

set -eou pipefail
set -xv

JOB_NAME=job.batch/cyclonus
JOB_NS=netpol
CLUSTER_NAME=kind-cilium

CLUSTER=$CLUSTER_NAME ./setup-kind.sh

docker pull mfenwick100/cyclonus:latest
kind load docker-image mfenwick100/cyclonus:latest --name $CLUSTER_NAME
#
docker pull k8s.gcr.io/e2e-test-images/agnhost:2.28
kind load docker-image k8s.gcr.io/e2e-test-images/agnhost:2.28 --name $CLUSTER_NAME

# get some debug info
kubectl get nodes
kubectl get pods -A

../run-cyclonus-job.sh ./cyclonus-job-github-action.yaml

# wait for job to start running
# TODO there's got to be a better way to do this
sleep 30
kubectl get all -A

kubectl wait --for=condition=ready pod -l job-name=cyclonus -n netpol --timeout=5m

kubectl logs -f -n $JOB_NS $JOB_NAME
