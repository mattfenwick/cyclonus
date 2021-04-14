#!/usr/bin/env sh

set -xv
set -eou pipefail

CYCLONUS_ARGS=$@
RESULTS_DIR="${RESULTS_DIR:-/tmp/results}"


./cyclonus $CYCLONUS_ARGS > "${RESULTS_DIR}"/results.txt


cd "${RESULTS_DIR}"

  # Sonobuoy worker expects a tar file.
tar czf results.tar.gz results.txt

# Signal to the worker that we are done and where to find the results.
realpath results.tar.gz > ./done
