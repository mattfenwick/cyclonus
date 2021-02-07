package generator

import (
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	networkingv1 "k8s.io/api/networking/v1"
)

type CreatePolicyAction struct {
	Policy *networkingv1.NetworkPolicy
}

func CreatePolicy(policy *networkingv1.NetworkPolicy) *Action {
	return &Action{CreatePolicy: &CreatePolicyAction{Policy: policy}}
}

type DeletePolicyAction struct {
	Namespace string
	Name      string
}

func DeletePolicy(ns string, name string) *Action {
	return &Action{DeletePolicy: &DeletePolicyAction{Namespace: ns, Name: name}}
}

type UpdatePolicyAction struct {
	Policy *networkingv1.NetworkPolicy
}

func UpdatePolicy(policy *networkingv1.NetworkPolicy) *Action {
	return &Action{UpdatePolicy: &UpdatePolicyAction{Policy: policy}}
}

type SetNamespaceLabelsAction struct {
	Namespace string
	Labels    map[string]string
}

func SetNamespaceLabels(ns string, labels map[string]string) *Action {
	return &Action{SetNamespaceLabels: &SetNamespaceLabelsAction{Namespace: ns, Labels: labels}}
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

type ReadNetworkPoliciesAction struct {
	Namespaces []string
}

func ReadNetworkPolicies(namespaces []string) *Action {
	return &Action{ReadNetworkPolicies: &ReadNetworkPoliciesAction{Namespaces: namespaces}}
}

// Action: exactly one field must be non-null.  This models a discriminated union (sum type).
type Action struct {
	CreatePolicy        *CreatePolicyAction
	UpdatePolicy        *UpdatePolicyAction
	DeletePolicy        *DeletePolicyAction
	SetNamespaceLabels  *SetNamespaceLabelsAction
	SetPodLabels        *SetPodLabelsAction
	ReadNetworkPolicies *ReadNetworkPoliciesAction
	// TODO create pod?  create namespace?
}

// ProbeConfig: exactly one field must be non-null (or, in AllAvailable's case, non-false).  This
//   models a discriminated union (sum type).
type ProbeConfig struct {
	AllAvailable bool
	PortProtocol *matcher.PortProtocol
}

type TestStep struct {
	Probe   *ProbeConfig
	Actions []*Action
}

func NewTestStep(pp *ProbeConfig, actions ...*Action) *TestStep {
	return &TestStep{
		Probe:   pp,
		Actions: actions,
	}
}

type TestCase struct {
	Description string
	Steps       []*TestStep
}

func NewSingleStepTestCase(description string, pp *ProbeConfig, actions ...*Action) *TestCase {
	return &TestCase{
		Description: description,
		Steps:       []*TestStep{NewTestStep(pp, actions...)},
	}
}

func NewTestCase(description string, steps ...*TestStep) *TestCase {
	return &TestCase{
		Description: description,
		Steps:       steps,
	}
}
