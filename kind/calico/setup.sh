#!/usr/bin/env bash

source ../lib/common.sh
source ../lib/calico.sh

trap calico_cleanup_cluster INT TERM

calico_create_cluster $(pwd)/conf.yaml

calico_apply_manifests
calico_set_node_options

echo "Waiting for calico to start running now... "
wait_all_pods_status_running 5

calico_download_client $(pwd)/calicoctl

# Show information about client and server
echo -e "\n==========================="
echo -e "\e[1m\e[32mCalico client output\e[0m"
echo -e "==========================="
./calicoctl version
