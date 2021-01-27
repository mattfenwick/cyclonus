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

// Action: exactly one field must be non-null
type Action struct {
	CreatePolicy *CreatePolicyAction
	// TODO uncomment these
	//DeletePolicy *DeletePolicyAction
	//UpdateNamespaceLabel *UpdateNamespaceLabelAction
	//RemoveNamespaceLabel *RemoveNamespaceLabelAction
	//RemovePodLabel *RemovePodLabelAction
	UpdatePodLabel *UpdatePodLabelAction
	// TODO create pod?  create namespace?
}

type TestCase struct {
	Actions []*Action
}
