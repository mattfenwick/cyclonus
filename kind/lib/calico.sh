#!/usr/bin/env bash

DIR_LIBS="$(dirname "${BASH_SOURCE[0]}")"
source "${DIR_LIBS}/kind.sh"

CLUSTER_NAME="netpol-calico"
CALICO_CLIENT_VERSION="v3.18.1"
NODE_OPTIONS=(FELIX_IGNORELOOSERPF=true FELIX_XDPENABLED=false)
MANIFESTS=(https://docs.projectcalico.org/manifests/calico.yaml)

function calico_set_node_options() {
    ##################################################################
    # Description:                                                   #
    #   Set Node Options                                             #
    ##################################################################
    for opt in "${NODE_OPTIONS[@]}"; do
        echo "Setting ${opt} to calico-node..."
        kubectl -n kube-system set env daemonset/calico-node "${opt}" 1> /dev/null
        if (( $? )) ; then
            echo -e "[ \e[1m\e[31mFAIL\e[0m  ] cannot to set ${opt} into the calico node"
            calico_cleanup_cluster "${CLUSTER_NAME}"
            exit 1
        fi
    done
}

function calico_apply_manifests() {
    ##################################################################
    # Description:                                                   #
    #   Apply any required manifest to calico                        #
    ##################################################################
    echo
    echo "Applying calico manifest..." 1> /dev/null
    for manifest in "${MANIFESTS[@]}"; do
        kubectl apply -f "${manifest}" 1> /dev/null
        if (( $? )) ; then
            echo -e "[ \e[1m\e[31mFAIL\e[0m  ] cannot apply ${manifest}"
            calico_cleanup_cluster "${CLUSTER_NAME}"
            exit 1
        fi
    done
}

function calico_download_client() {
    ##################################################################
    # Description:                                                   #
    #   Download calicoctl and make the file executable              #
    ##################################################################
    echo
    echo "Downloading calicoctl.."
    path_calicoctl="${1}"
    CALICO_CLIENT_VERSION="v3.18.1"
    curl -L https://github.com/projectcalico/calicoctl/releases/download/"${CALICO_CLIENT_VERSION}"/calicoctl --output "${path_calicoctl}"
    if (( $? )) ; then
        echo -e "[ \e[1m\e[31mFAIL\e[0m  ] cannot download calicoctl ${CALICO_VERSION}"
        exit 1
    fi
    chmod +x "${path_calicoctl}"
}

function calico_cleanup_cluster() {
    ##################################################################
    # Description:                                                   #
    #   cleanup the kind cluster                                     #
    ##################################################################
    kind_cleanup_cluster "${CLUSTER_NAME}"
}

function calico_create_cluster() {
    ##################################################################
    # Description:                                                   #
    #   create the kind cluster                                      #
    ##################################################################
    kind_create_cluster "${CLUSTER_NAME}"
}
