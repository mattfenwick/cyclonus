#!/usr/bin/env bash

set -eou pipefail
set -xv

CLUSTER_NAME=kind-cilium

CLUSTER=$CLUSTER_NAME ./setup-kind.sh

JOB_YAML=./cyclonus-job-github-action.yaml CLUSTER_NAME=$CLUSTER_NAME ../run-cyclonus-job.sh
