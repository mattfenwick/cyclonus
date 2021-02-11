package generator

import (
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
	PortProtocol *PortProtocol
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
	Features    *Features
	Steps       []*TestStep
}

func NewSingleStepTestCase(description string, pp *ProbeConfig, actions ...*Action) *TestCase {
	return &TestCase{
		Description: description,
		Features:    nil,
		Steps:       []*TestStep{NewTestStep(pp, actions...)},
	}
}

func NewTestCase(description string, steps ...*TestStep) *TestCase {
	return &TestCase{
		Description: description,
		Steps:       steps,
	}
}

func (t *TestCase) GetFeatures() *Features {
	derived := t.DerivedFeatures()
	if t.Features != nil {
		features := t.Features
		features.Action = mergeActionFeatureSets(features.Action, derived.Action)
		return features
	}
	return derived
}

//func (t *TestCase) SortedFeatures() []Feature {
//	var slice []Feature
//	features := t.Features()
//	for f := range features {
//		slice = append(slice, f)
//	}
//	sort.Slice(slice, func(i, j int) bool {
//		return slice[i] < slice[j]
//	})
//	return slice
//}

func (t *TestCase) DerivedFeatures() *Features {
	features := &Features{}
	for _, step := range t.Steps {
		for _, action := range step.Actions {
			var policy *networkingv1.NetworkPolicy
			actionFeatures := map[ActionFeature]bool{}
			if action.DeletePolicy != nil {
				actionFeatures[ActionFeatureDeletePolicy] = true
			} else if action.ReadNetworkPolicies != nil {
				// TODO need to also analyze these policies after they get read
				actionFeatures[ActionFeatureReadPolicies] = true
			} else if action.SetPodLabels != nil {
				actionFeatures[ActionFeatureSetPodLabels] = true
			} else if action.SetNamespaceLabels != nil {
				actionFeatures[ActionFeatureSetNamespaceLabels] = true
			} else if action.UpdatePolicy != nil {
				actionFeatures[ActionFeatureUpdatePolicy] = true
				policy = action.UpdatePolicy.Policy
			} else if action.CreatePolicy != nil {
				actionFeatures[ActionFeatureCreatePolicy] = true
				policy = action.CreatePolicy.Policy
			} else {
				panic("invalid Action")
			}
			newFeatures := &Features{Action: actionFeatures}
			if policy != nil {
				newFeatures = newFeatures.Combine(GetFeaturesForPolicy(policy))
			}
			features = features.Combine(newFeatures)
		}
	}
	return features
}
