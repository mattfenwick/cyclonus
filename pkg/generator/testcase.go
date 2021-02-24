package generator

import (
	networkingv1 "k8s.io/api/networking/v1"
)

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

func (t *TestCase) GetFeatures() map[string]bool {
	return t.DerivedFeatures()
}

func (t *TestCase) GetFeaturesSlice() []string {
	var features []string
	for f := range t.GetFeatures() {
		features = append(features, f)
	}
	return features
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

func (t *TestCase) DerivedFeatures() map[string]bool {
	features := map[string]bool{}
	for _, step := range t.Steps {
		for _, action := range step.Actions {
			var policy *networkingv1.NetworkPolicy
			actionFeatures := map[string]bool{}
			if action.CreatePolicy != nil {
				actionFeatures[ActionFeatureCreatePolicy] = true
				policy = action.CreatePolicy.Policy
			} else if action.UpdatePolicy != nil {
				actionFeatures[ActionFeatureUpdatePolicy] = true
				policy = action.UpdatePolicy.Policy
			} else if action.DeletePolicy != nil {
				actionFeatures[ActionFeatureDeletePolicy] = true
			} else if action.CreateNamespace != nil {
				actionFeatures[ActionFeatureCreateNamespace] = true
			} else if action.SetNamespaceLabels != nil {
				actionFeatures[ActionFeatureSetNamespaceLabels] = true
			} else if action.DeleteNamespace != nil {
				actionFeatures[ActionFeatureDeleteNamespace] = true
			} else if action.ReadNetworkPolicies != nil {
				// TODO need to also analyze these policies after they get read
				actionFeatures[ActionFeatureReadPolicies] = true
			} else if action.CreatePod != nil {
				actionFeatures[ActionFeatureCreatePod] = true
			} else if action.SetPodLabels != nil {
				actionFeatures[ActionFeatureSetPodLabels] = true
			} else if action.DeletePod != nil {
				actionFeatures[ActionFeatureDeletePod] = true
			} else {
				panic("invalid Action")
			}

			features = mergeSets(features, actionFeatures)
			if policy != nil {
				features = mergeSets(features, GetFeaturesForPolicy(NewNetpol(policy)))
			}
		}
	}
	return features
}
