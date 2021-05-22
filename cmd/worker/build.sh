#!/usr/bin/env bash

set -euo pipefail
set -xv

CLUSTER=${CLUSTER:-netpol-calico}
IMAGE=mfenwick100/cyclonus-worker:latest

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o worker ./main.go

docker build -t $IMAGE .

kind load docker-image $IMAGE --name "$CLUSTER"
