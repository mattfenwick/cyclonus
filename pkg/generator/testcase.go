package generator

import (
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sort"
	"strings"
)

type TestCase struct {
	Description string
	Tags        StringSet
	Steps       []*TestStep
}

func NewSingleStepTestCase(description string, tags StringSet, pp *ProbeConfig, actions ...*Action) *TestCase {
	if description == "" {
		tagSlice := tags.Keys()
		sort.Strings(tagSlice)
		description = strings.Join(tagSlice, ",")
	}
	return &TestCase{
		Description: description,
		Tags:        tags,
		Steps:       []*TestStep{NewTestStep(pp, actions...)},
	}
}

func NewTestCase(description string, tags StringSet, steps ...*TestStep) *TestCase {
	return &TestCase{
		Description: description,
		Tags:        tags,
		Steps:       steps,
	}
}

func (t *TestCase) collectActionsAndPolicies() (map[string]bool, []*networkingv1.NetworkPolicy) {
	features := map[string]bool{}
	var policies []*networkingv1.NetworkPolicy
	for _, step := range t.Steps {
		for _, action := range step.Actions {
			if action.CreatePolicy != nil {
				features[ActionFeatureCreatePolicy] = true
				policies = append(policies, action.CreatePolicy.Policy)
			} else if action.UpdatePolicy != nil {
				features[ActionFeatureUpdatePolicy] = true
				policies = append(policies, action.UpdatePolicy.Policy)
			} else if action.DeletePolicy != nil {
				features[ActionFeatureDeletePolicy] = true
			} else if action.CreateNamespace != nil {
				features[ActionFeatureCreateNamespace] = true
			} else if action.SetNamespaceLabels != nil {
				features[ActionFeatureSetNamespaceLabels] = true
			} else if action.DeleteNamespace != nil {
				features[ActionFeatureDeleteNamespace] = true
			} else if action.ReadNetworkPolicies != nil {
				// TODO need to also analyze these policies after they get read
				features[ActionFeatureReadPolicies] = true
			} else if action.CreatePod != nil {
				features[ActionFeatureCreatePod] = true
			} else if action.SetPodLabels != nil {
				features[ActionFeatureSetPodLabels] = true
			} else if action.DeletePod != nil {
				features[ActionFeatureDeletePod] = true
			} else {
				panic("invalid Action")
			}
		}
	}
	return features, policies
}

func (t *TestCase) GetFeatures() map[string][]string {
	actionSet, policies := t.collectActionsAndPolicies()
	generalSet, ingressSet, egressSet := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, policy := range policies {
		parsedPolicy := NewNetpol(policy)
		generalSet = mergeSets(generalSet, GeneralNetpolTraverser.Traverse(parsedPolicy))
		ingressSet = mergeSets(ingressSet, IngressNetpolTraverser.Traverse(parsedPolicy))
		egressSet = mergeSets(egressSet, EgressNetpolTraverser.Traverse(parsedPolicy))
	}
	return map[string][]string{
		"general": setToSlice(generalSet),
		"ingress": setToSlice(ingressSet),
		"egress":  setToSlice(egressSet),
		"action":  setToSlice(actionSet),
	}
}

func setToSlice(set map[string]bool) []string {
	var slice []string
	for f := range set {
		slice = append(slice, f)
	}
	return slice
}

func mergeSets(l, r map[string]bool) map[string]bool {
	merged := map[string]bool{}
	for k := range l {
		merged[k] = true
	}
	for k := range r {
		merged[k] = true
	}
	return merged
}

// ProbeConfig: exactly one field must be non-null (or, in AllAvailable's case, non-false).  This
//   models a discriminated union (sum type).
type ProbeConfig struct {
	AllAvailable bool
	PortProtocol *PortProtocol
}

type PortProtocol struct {
	Protocol v1.Protocol
	Port     intstr.IntOrString
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
