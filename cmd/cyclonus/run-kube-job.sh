#!/bin/bash

#set -euo pipefail
set -xv

kubectl create clusterrolebinding cyclonus --clusterrole=cluster-admin --serviceaccount=kube-system:cyclonus
kubectl create sa cyclonus -n kube-system

kubectl create -f cyclonus-job.yaml
