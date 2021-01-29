package generator

import (
	v1 "k8s.io/api/core/v1"
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
	Port     int
	Protocol v1.Protocol
	Actions  []*Action
}

type TestCase struct {
	Description string
	Steps       []*TestStep
}

func NewTestCase(description string, port int, protocol v1.Protocol, actions []*Action) *TestCase {
	return &TestCase{
		Description: description,
		Steps: []*TestStep{
			{
				Port:     port,
				Protocol: protocol,
				Actions:  actions,
			},
		}}
}
