#!/usr/bin/env bash

set -euo pipefail
set -xv

# run all 4
go run ../cmd/cyclonus/main.go analyze \
  --use-example-policies \
  --target-pod-path ./targets-example.json \
  --traffic-path ./traffic-example.json \
  --probe-path ./synthetic-probe-example.json

# run just the explainer
go run ../cmd/cyclonus/main.go analyze \
  --policy-path ../networkpolicies/simple-example/

# run just the targets
go run ../cmd/cyclonus/main.go analyze \
  --explain=false \
  --use-example-policies \
  --target-pod-path ./targets-example.json

# run just the traffic
go run ../cmd/cyclonus/main.go analyze \
  --explain=false \
  --use-example-policies \
  --traffic-path ./traffic-example.json

# run just the probe
go run ../cmd/cyclonus/main.go analyze \
  --explain=false \
  --use-example-policies \
  --probe-path ./synthetic-probe-example.json