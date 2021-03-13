#!/bin/bash

#set -euo pipefail
set -xv

NS=${NS:-netpol}
JOB=$1

kubectl create ns "$NS"
kubectl create clusterrolebinding cyclonus --clusterrole=cluster-admin --serviceaccount="$NS":cyclonus
kubectl create sa cyclonus -n "$NS"

kubectl create -f $1 -n "$NS"
