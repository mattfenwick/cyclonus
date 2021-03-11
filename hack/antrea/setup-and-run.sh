#!/bin/bash

set -euo pipefail
set -xv

./setup-kind.sh
../run-cyclonus-job.sh cyclonus-job.yaml
