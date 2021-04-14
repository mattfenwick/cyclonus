#!/usr/bin/env sh

set -xv
set -eou pipefail

CYCLONUS_ARGS=${CYCLONUS_ARGS:-"generate --include conflict --exclude egress,direction"}
results_dir="${RESULTS_DIR:-/tmp/results}"


./cyclonus $CYCLONUS_ARGS > "${results_dir}"/results.txt


cd "${results_dir}"

  # Sonobuoy worker expects a tar file.
tar czf results.tar.gz results.txt

# Signal to the worker that we are done and where to find the results.
printf "${results_dir}"/results.tar.gz > "${results_dir}"/done
