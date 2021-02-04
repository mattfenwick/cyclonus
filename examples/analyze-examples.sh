#!/usr/bin/env bash

set -euo pipefail
set -xv

go run ../cmd/cyclonus/main.go analyze \
  --use-example-policies \
  --target-pod-path ./targets-example.json \
  --traffic-path ./traffic-example.json \
  --probe-path ./synthetic-probe-example.json
