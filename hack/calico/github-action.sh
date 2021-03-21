#!/usr/bin/env bash

set -eou pipefail
set -xv

CLUSTER_NAME=kind-calico

CLUSTER=$CLUSTER_NAME ./setup-kind.sh

JOB_YAML=./cyclonus-job.yaml CLUSTER_NAME=$CLUSTER_NAME ../run-cyclonus-job.sh
