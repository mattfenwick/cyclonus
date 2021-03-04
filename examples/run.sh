#!/usr/bin/env bash

set -euo pipefail
set -xv

# run all 5
go run ../cmd/cyclonus/main.go analyze \
  --use-example-policies \
  --mode explain,lint,query-target,query-traffic,probe \
  --target-pod-path ./targets-example.json \
  --traffic-path ./traffic-example.json \
  --probe-path ./probe-example.json

# run just the explainer
go run ../cmd/cyclonus/main.go analyze \
  --mode explain \
  --policy-path ../networkpolicies/simple-example/

# run just the targets
go run ../cmd/cyclonus/main.go analyze \
  --mode query-target \
  --policy-path ../networkpolicies/simple-example/ \
  --target-pod-path ./targets.json

# run just the traffic
go run ../cmd/cyclonus/main.go analyze \
  --mode query-traffic \
  --policy-path ../networkpolicies/simple-example/ \
  --traffic-path ./traffic.json

# run just the probe
go run ../cmd/cyclonus/main.go analyze \
  --mode probe \
  --policy-path ../networkpolicies/simple-example/ \
  --probe-path ./probe.json

# run just the linter
go run ../cmd/cyclonus/main.go analyze \
  --mode lint \
  --policy-path ../networkpolicies/simple-example