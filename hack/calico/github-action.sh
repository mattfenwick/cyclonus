#!/usr/bin/env bash

set -eo pipefail
set -xv

WAIT_TIMEOUT=240m
JOB_NAME=job.batch/cyclonus
JOB_NS=netpol
CLUSTER_NAME=kind-calico

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

# wait for job to come up
kubectl get pods -n netpol
sleep 5
kubectl get pods -n netpol

kubectl wait --for=condition=ready pod -l job-name=cyclonus -n netpol

kubectl logs -f -n $JOB_NS $JOB_NAME
