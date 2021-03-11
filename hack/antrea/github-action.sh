#!/usr/bin/env bash

set -eo pipefail
set -xv

WAIT_TIMEOUT=240m
JOB_NAME=job.batch/cyclonus
JOB_NS=netpol

./setup-kind.sh

# preload images
kind load docker-image projects.registry.vmware.com/antrea/antrea-ubuntu:latest
#
docker pull mfenwick100/cyclonus:latest
kind load docker-image mfenwick100/cyclonus:latest
#
docker pull k8s.gcr.io/e2e-test-images/agnhost:2.28
kind load docker-image k8s.gcr.io/e2e-test-images/agnhost:2.28

# get some debug info
kubectl get nodes
kubectl get pods -A

../run-cyclonus-job.sh ./cyclonus-job.yaml

time kubectl wait --for=condition=complete --timeout=$WAIT_TIMEOUT -n $JOB_NS $JOB_NAME

echo "===> Checking cyclonus results <==="

LOG_FILE=$(mktemp)
kubectl logs -n $JOB_NS $JOB_NAME > "$LOG_FILE"
cat "$LOG_FILE"

rc=0
cat "$LOG_FILE" | grep "failure" > /dev/null 2>&1 || rc=$?
if [ $rc -eq 0 ]; then
    exit 1
fi
