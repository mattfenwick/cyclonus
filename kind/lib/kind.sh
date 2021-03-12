#!/usr/bin/env bash

function kind_create_cluster() {
    ##################################################################
    # Description:                                                   #
    #   Create cluster via kind                                      #
    ##################################################################

    cluster_name="${1}"
    conf="${2}"

    kind create cluster --name "${cluster_name}" --config "${conf}"
    if (( $? )) ; then
        echo -e "[ \e[1m\e[31mFAIL\e[0m  ] cannot create cluster ${cluster_name} ${conf}"
        exit 1
    fi

    until kubectl cluster-info;  do
        echo "$(date)waiting for cluster..."
        sleep 2
    done
}


function kind_cleanup_cluster() {
    ##################################################################
    # Description:                                                   #
    #   cleanup the kind cluster                                     #
    ##################################################################

    cluster_name="${1}"

    echo "Cleanup cluster ${cluster_name}..."
    kind delete cluster --name "${cluster_name}"
    if (( $? )) ; then
        echo -e "[ \e[1m\e[31mFAIL\e[0m  ] cannot cleanup cluster ${cluster_name}"
        exit 1
    fi
}
