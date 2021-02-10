#!/usr/bin/env bash

set -euo pipefail
set -xv

# run all 5
go run ../cmd/cyclonus/main.go analyze \
  --use-example-policies \
  --lint=true \
  --target-pod-path ./targets-example.json \
  --traffic-path ./traffic-example.json \
  --probe-path ./probe-example.json

# run just the explainer
go run ../cmd/cyclonus/main.go analyze \
  --policy-path ../networkpolicies/simple-example/

# run just the targets
go run ../cmd/cyclonus/main.go analyze \
  --explain=false \
  --policy-path ../networkpolicies/simple-example/ \
  --target-pod-path ./targets.json

# run just the traffic
go run ../cmd/cyclonus/main.go analyze \
  --explain=false \
  --policy-path ../networkpolicies/simple-example/ \
  --traffic-path ./traffic.json

# run just the probe
go run ../cmd/cyclonus/main.go analyze \
  --explain=false \
  --policy-path ../networkpolicies/simple-example/ \
  --probe-path ./probe.json

# run just the linter
go run ../cmd/cyclonus/main.go analyze \
  --explain=false \
  --lint=true \
  --policy-path ../networkpolicies/simple-example