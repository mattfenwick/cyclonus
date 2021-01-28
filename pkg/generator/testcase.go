package generator

import (
	networkingv1 "k8s.io/api/networking/v1"
)

type CreatePolicyAction struct {
	Policy *networkingv1.NetworkPolicy
}

type DeletePolicyAction struct {
	Policy *networkingv1.NetworkPolicy
}

type UpdateNamespaceLabelAction struct {
	Namespace string
	Key       string
	Value     string
}

type RemoveNamespaceLabelAction struct {
	Namespace string
	Key       string
}

type UpdatePodLabelAction struct {
	Namespace string
	Pod       string
	Key       string
	Value     string
}

type RemovePodLabelAction struct {
	Namespace string
	Pod       string
	Key       string
}

type ReadNetworkPoliciesAction struct {
	Namespaces []string
}

func CreatePolicy(policy *networkingv1.NetworkPolicy) *Action {
	return &Action{CreatePolicy: &CreatePolicyAction{Policy: policy}}
}

func UpdatePodLabel(namespace string, pod string, key string, value string) *Action {
	return &Action{UpdatePodLabel: &UpdatePodLabelAction{
		Namespace: namespace,
		Pod:       pod,
		Key:       key,
		Value:     value,
	}}
}

func ReadNetworkPolicies(namespaces []string) *Action {
	return &Action{ReadNetworkPolicies: &ReadNetworkPoliciesAction{Namespaces: namespaces}}
}

// Action: exactly one field must be non-null
type Action struct {
	CreatePolicy *CreatePolicyAction
	// TODO uncomment these
	//DeletePolicy *DeletePolicyAction
	//UpdateNamespaceLabel *UpdateNamespaceLabelAction
	//RemoveNamespaceLabel *RemoveNamespaceLabelAction
	//RemovePodLabel *RemovePodLabelAction
	UpdatePodLabel      *UpdatePodLabelAction
	ReadNetworkPolicies *ReadNetworkPoliciesAction
	// TODO create pod?  create namespace?
}

type TestStep struct {
	Actions []*Action
}

type TestCase struct {
	Steps []*TestStep
}

func NewTestCase(actions []*Action) *TestCase {
	return &TestCase{Steps: []*TestStep{
		{
			Actions: actions,
		},
	}}
}
