#!/usr/bin/env bash

function cleanup_cluster() {
    ##################################################################
    # Description:                                                   #
    #   Destroy the kind cluster                                     #
    #                                                                #
    # Args:                                                          #
    #     ${1} = cluster name                                        #
    ##################################################################
    cluster_name="{$1}"
    kind delete cluster --name "${cluster_name}"
} 

function wait_all_pods_status_running() {
    ##################################################################
    # Description:                                                   #
    #   Wait all pods in the cluster be in running status            #
    #                                                                #
    # Args:                                                          #
    #     ${1} = seconds to be used to wait until the next kubectl   #
    #            command                                             #
    ##################################################################
    sleep_sec="${1}"
    while true ; do
        cmd=$(kubectl get pods -A --field-selector=status.phase!=Running -o name)
        if [ -z "${cmd}" ];
        then
           break
        fi
        echo
        numberpods=$(echo "${cmd}" | wc -l)
        echo -e "[\e[1m\e[32m$(date +"%r")\e[0m] The following \e[1m\e[32m${numberpods} pods are NOT in RUNNING status yet\e[0m... please wait $sleep_sec seconds for a auto-refresh.."
        echo -e "${cmd}"
        sleep "${sleep_sec}"
    done
}
