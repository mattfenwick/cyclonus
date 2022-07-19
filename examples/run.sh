#!/usr/bin/env bash

set -euo pipefail
set -xv

CYCLONUS_OUTPUT_DIR=${CYCLONUS_OUTPUT_DIR:-"./cyclonus-output"}

mkdir -p "$CYCLONUS_OUTPUT_DIR"

# run all 5
printf "\n\n********************** run all 5 modes **************************\n\n"
go run ../cmd/cyclonus/main.go analyze \
  --use-example-policies \
  --mode explain,lint,query-target,query-traffic,probe \
  --target-pod-path ./targets-example.json \
  --traffic-path ./traffic-example.json \
  --probe-path ./probe-example.json \
  > "$CYCLONUS_OUTPUT_DIR"/analyze-all-five.txt

# run just the explainer
printf "\n\n********************** run just the explainer **************************\n\n"
go run ../cmd/cyclonus/main.go analyze \
  --mode explain \
  --policy-path ../networkpolicies/simple-example/ \
  > "$CYCLONUS_OUTPUT_DIR"/analyze-explain.txt

# run just the targets
printf "\n\n********************** run just the targets **************************\n\n"
go run ../cmd/cyclonus/main.go analyze \
  --mode query-target \
  --policy-path ../networkpolicies/simple-example/ \
  --target-pod-path ./targets.json \
  > "$CYCLONUS_OUTPUT_DIR"/analyze-query-target.txt

# run just the traffic
printf "\n\n********************** run just the traffic **************************\n\n"
go run ../cmd/cyclonus/main.go analyze \
  --mode query-traffic \
  --policy-path ../networkpolicies/simple-example/ \
  --traffic-path ./traffic.json \
  > "$CYCLONUS_OUTPUT_DIR"/analyze-query-traffic.txt

# run just the probe
printf "\n\n********************** run just the probe **************************\n\n"
go run ../cmd/cyclonus/main.go analyze \
  --mode probe \
  --policy-path ../networkpolicies/simple-example/ \
  --probe-path ./probe.json \
  > "$CYCLONUS_OUTPUT_DIR"/analyze-probe.txt

# run just the linter
printf "\n\n********************** run just the linter **************************\n\n"
go run ../cmd/cyclonus/main.go analyze \
  --mode lint \
  --policy-path ../networkpolicies/simple-example \
  > "$CYCLONUS_OUTPUT_DIR"/analyze-lint.txt
