package generator

import networkingv1 "k8s.io/api/networking/v1"

// Action models a sum type (discriminated union): exactly one field must be non-null.
type Action struct {
	CreatePolicy *CreatePolicyAction
	UpdatePolicy *UpdatePolicyAction
	DeletePolicy *DeletePolicyAction

	CreateNamespace    *CreateNamespaceAction
	SetNamespaceLabels *SetNamespaceLabelsAction
	DeleteNamespace    *DeleteNamespaceAction

	ReadNetworkPolicies *ReadNetworkPoliciesAction

	CreatePod    *CreatePodAction
	SetPodLabels *SetPodLabelsAction
	DeletePod    *DeletePodAction
}

type CreatePolicyAction struct {
	Policy *networkingv1.NetworkPolicy
}

func CreatePolicy(policy *networkingv1.NetworkPolicy) *Action {
	return &Action{CreatePolicy: &CreatePolicyAction{Policy: policy}}
}

type UpdatePolicyAction struct {
	Policy *networkingv1.NetworkPolicy
}

func UpdatePolicy(policy *networkingv1.NetworkPolicy) *Action {
	return &Action{UpdatePolicy: &UpdatePolicyAction{Policy: policy}}
}

type DeletePolicyAction struct {
	Namespace string
	Name      string
}

func DeletePolicy(ns string, name string) *Action {
	return &Action{DeletePolicy: &DeletePolicyAction{Namespace: ns, Name: name}}
}

type CreateNamespaceAction struct {
	Namespace string
	Labels    map[string]string
}

func CreateNamespace(ns string, labels map[string]string) *Action {
	return &Action{CreateNamespace: &CreateNamespaceAction{Namespace: ns, Labels: labels}}
}

type SetNamespaceLabelsAction struct {
	Namespace string
	Labels    map[string]string
}

func SetNamespaceLabels(ns string, labels map[string]string) *Action {
	return &Action{SetNamespaceLabels: &SetNamespaceLabelsAction{Namespace: ns, Labels: labels}}
}

type DeleteNamespaceAction struct {
	Namespace string
}

func DeleteNamespace(ns string) *Action {
	return &Action{DeleteNamespace: &DeleteNamespaceAction{Namespace: ns}}
}

type ReadNetworkPoliciesAction struct {
	Namespaces []string
}

func ReadNetworkPolicies(namespaces []string) *Action {
	return &Action{ReadNetworkPolicies: &ReadNetworkPoliciesAction{Namespaces: namespaces}}
}

type CreatePodAction struct {
	Namespace string
	Pod       string
	Labels    map[string]string
}

func CreatePod(namespace string, pod string, labels map[string]string) *Action {
	return &Action{CreatePod: &CreatePodAction{
		Namespace: namespace,
		Pod:       pod,
		Labels:    labels,
	}}
}

type SetPodLabelsAction struct {
	Namespace string
	Pod       string
	Labels    map[string]string
}

func SetPodLabels(namespace string, pod string, labels map[string]string) *Action {
	return &Action{SetPodLabels: &SetPodLabelsAction{
		Namespace: namespace,
		Pod:       pod,
		Labels:    labels,
	}}
}

type DeletePodAction struct {
	Namespace string
	Pod       string
}

func DeletePod(namespace string, pod string) *Action {
	return &Action{DeletePod: &DeletePodAction{
		Namespace: namespace,
		Pod:       pod,
	}}
}
