package connectivity

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sort"
)

type Feature string

const (
	FeatureCreatePolicy       Feature = "CreatePolicy"
	FeatureDeletePolicy       Feature = "DeletePolicy"
	FeatureReadPolicies       Feature = "ReadPolicies"
	FeatureSetPodLabels       Feature = "SetPodLabels"
	FeatureSetNamespaceLabels Feature = "SetNamespaceLabels"
	FeatureUpdatePolicy       Feature = "UpdatePolicy"

	FeatureTargetPodSelectorEmpty            Feature = "TargetPodSelectorEmpty"
	FeatureTargetPodSelectorMatchLabels      Feature = "TargetPodSelectorMatchLabels"
	FeatureTargetPodSelectorMatchExpressions Feature = "TargetPodSelectorMatchExpressions"

	FeatureIngressEmptyPortSlice Feature = "IngressEmptyPortSlice"
	FeatureIngressNumberedPort   Feature = "IngressNumberedPort"
	FeatureIngressNamedPort      Feature = "IngressNamedPort"
	FeatureIngressNilPort        Feature = "IngressNilPort"

	FeatureEgressEmptyPortSlice Feature = "EgressEmptyPortSlice"
	FeatureEgressNumberedPort   Feature = "EgressNumberedPort"
	FeatureEgressNamedPort      Feature = "EgressNamedPort"
	FeatureEgressNilPort        Feature = "EgressNilPort"

	FeatureIngressEmptyPeerSlice                        Feature = "IngressEmptyPeerSlice"
	FeatureIngressPeerIPBlock                           Feature = "IngressPeerIPBlock"
	FeatureIngressPeerIPBlockEmptyExcept                Feature = "IngressPeerIPBlockEmptyExcept"
	FeatureIngressPeerIPBlockNonemptyExcept             Feature = "IngressPeerIPBlockNonemptyExcept"
	FeatureIngressPeerPodSelectorEmpty                  Feature = "IngressPeerPodSelectorEmpty"
	FeatureIngressPeerPodSelectorMatchLabels            Feature = "IngressPeerPodSelectorMatchLabels"
	FeatureIngressPeerPodSelectorMatchExpressions       Feature = "IngressPeerPodSelectorMatchExpressions"
	FeatureIngressPeerPodSelectorNil                    Feature = "IngressPeerPodSelectorNil"
	FeatureIngressPeerNamespaceSelectorEmpty            Feature = "IngressPeerNamespaceSelectorEmpty"
	FeatureIngressPeerNamespaceSelectorMatchLabels      Feature = "IngressPeerNamespaceSelectorMatchLabels"
	FeatureIngressPeerNamespaceSelectorMatchExpressions Feature = "IngressPeerNamespaceSelectorMatchExpressions"
	FeatureIngressPeerNamespaceSelectorNil              Feature = "IngressPeerNamespaceSelectorNil"

	FeatureEgressEmptyPeerSlice                        Feature = "EgressEmptyPeerSlice"
	FeatureEgressPeerIPBlock                           Feature = "EgressPeerIPBlock"
	FeatureEgressPeerIPBlockEmptyExcept                Feature = "EgressPeerIPBlockEmptyExcept"
	FeatureEgressPeerIPBlockNonemptyExcept             Feature = "EgressPeerIPBlockNonemptyExcept"
	FeatureEgressPeerPodSelectorEmpty                  Feature = "EgressPeerPodSelectorEmpty"
	FeatureEgressPeerPodSelectorMatchLabels            Feature = "EgressPeerPodSelectorMatchLabels"
	FeatureEgressPeerPodSelectorMatchExpressions       Feature = "EgressPeerPodSelectorMatchExpressions"
	FeatureEgressPeerPodSelectorNil                    Feature = "EgressPeerPodSelectorNil"
	FeatureEgressPeerNamespaceSelectorEmpty            Feature = "EgressPeerNamespaceSelectorEmpty"
	FeatureEgressPeerNamespaceSelectorMatchLabels      Feature = "EgressPeerNamespaceSelectorMatchLabels"
	FeatureEgressPeerNamespaceSelectorMatchExpressions Feature = "EgressPeerNamespaceSelectorMatchExpressions"
	FeatureEgressPeerNamespaceSelectorNil              Feature = "EgressPeerNamespaceSelectorNil"

	FeatureIngressEmptySlice Feature = "IngressEmptySlice"
	FeatureEgressEmptySlice  Feature = "EgressEmptySlice"

	FeatureIngress          Feature = "Ingress"
	FeatureEgress           Feature = "Egress"
	FeatureIngressAndEgress Feature = "IngressAndEgress"
)

// AllFeatures is a slice of all Features.  Make sure to keep it updated as new Features are created.
var AllFeatures = []Feature{
	FeatureCreatePolicy,
	FeatureDeletePolicy,
	FeatureReadPolicies,
	FeatureSetPodLabels,
	FeatureSetNamespaceLabels,
	FeatureUpdatePolicy,

	FeatureTargetPodSelectorEmpty,
	FeatureTargetPodSelectorMatchLabels,
	FeatureTargetPodSelectorMatchExpressions,

	FeatureIngressEmptyPortSlice,
	FeatureIngressNumberedPort,
	FeatureIngressNamedPort,
	FeatureIngressNilPort,

	FeatureEgressEmptyPortSlice,
	FeatureEgressNumberedPort,
	FeatureEgressNamedPort,
	FeatureEgressNilPort,

	FeatureIngressEmptyPeerSlice,
	FeatureIngressPeerIPBlock,
	FeatureIngressPeerIPBlockEmptyExcept,
	FeatureIngressPeerIPBlockNonemptyExcept,
	FeatureIngressPeerPodSelectorEmpty,
	FeatureIngressPeerPodSelectorMatchLabels,
	FeatureIngressPeerPodSelectorMatchExpressions,
	FeatureIngressPeerPodSelectorNil,
	FeatureIngressPeerNamespaceSelectorEmpty,
	FeatureIngressPeerNamespaceSelectorMatchLabels,
	FeatureIngressPeerNamespaceSelectorMatchExpressions,
	FeatureIngressPeerNamespaceSelectorNil,

	FeatureEgressEmptyPeerSlice,
	FeatureEgressPeerIPBlock,
	FeatureEgressPeerIPBlockEmptyExcept,
	FeatureEgressPeerIPBlockNonemptyExcept,
	FeatureEgressPeerPodSelectorEmpty,
	FeatureEgressPeerPodSelectorMatchLabels,
	FeatureEgressPeerPodSelectorMatchExpressions,
	FeatureEgressPeerPodSelectorNil,
	FeatureEgressPeerNamespaceSelectorEmpty,
	FeatureEgressPeerNamespaceSelectorMatchLabels,
	FeatureEgressPeerNamespaceSelectorMatchExpressions,
	FeatureEgressPeerNamespaceSelectorNil,

	FeatureIngressEmptySlice,
	FeatureEgressEmptySlice,

	FeatureIngress,
	FeatureEgress,
	FeatureIngressAndEgress,
}

func (r *Result) SortedFeatures() []Feature {
	var slice []Feature
	features := r.Features()
	for f := range features {
		slice = append(slice, f)
	}
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})
	return slice
}

func (r *Result) Features() map[Feature]bool {
	features := map[Feature]bool{}
	for _, stepResult := range r.Steps {
		for _, policy := range stepResult.KubePolicies {
			features = mergeSets(features, specFeatures(policy.Spec))
		}
	}
	for _, step := range r.TestCase.Steps {
		for _, action := range step.Actions {
			if action.DeletePolicy != nil {
				features[FeatureDeletePolicy] = true
			} else if action.ReadNetworkPolicies != nil {
				features[FeatureReadPolicies] = true
			} else if action.SetPodLabels != nil {
				features[FeatureSetPodLabels] = true
			} else if action.SetNamespaceLabels != nil {
				features[FeatureSetNamespaceLabels] = true
			} else if action.UpdatePolicy != nil {
				features[FeatureUpdatePolicy] = true
			} else if action.CreatePolicy != nil {
				features[FeatureCreatePolicy] = true
			} else {
				panic("invalid Action")
			}
		}
	}
	return features
}

func specFeatures(spec networkingv1.NetworkPolicySpec) map[Feature]bool {
	features := targetPodSelectorFeatures(spec.PodSelector)
	ingress, egress := false, false
	for _, policyType := range spec.PolicyTypes {
		if policyType == networkingv1.PolicyTypeIngress {
			ingress = true
			features = mergeSets(features, ingressFeatures(spec.Ingress))
		} else if policyType == networkingv1.PolicyTypeEgress {
			egress = true
			features = mergeSets(features, egressFeatures(spec.Egress))
		}
	}
	if ingress {
		features[FeatureIngress] = true
	}
	if egress {
		features[FeatureEgress] = true
	}
	if ingress && egress {
		features[FeatureIngressAndEgress] = true
	}
	return features
}

func ingressFeatures(rules []networkingv1.NetworkPolicyIngressRule) map[Feature]bool {
	features := map[Feature]bool{}
	if len(rules) == 0 {
		features[FeatureIngressEmptySlice] = true
	} else {
		for _, rule := range rules {
			features = mergeSets(features, peerIngressFeatures(rule.From))
			features = mergeSets(features, portIngressFeatures(rule.Ports))
		}
	}
	return features
}

func egressFeatures(rules []networkingv1.NetworkPolicyEgressRule) map[Feature]bool {
	features := map[Feature]bool{}
	if len(rules) == 0 {
		features[FeatureEgressEmptySlice] = true
	} else {
		for _, rule := range rules {
			features = mergeSets(features, peerEgressFeatures(rule.To))
			features = mergeSets(features, portEgressFeatures(rule.Ports))
		}
	}
	return features
}

func peerIngressFeatures(peers []networkingv1.NetworkPolicyPeer) map[Feature]bool {
	features := map[Feature]bool{}
	if len(peers) == 0 {
		features[FeatureIngressEmptyPeerSlice] = true
	} else {
		for _, peer := range peers {
			if peer.IPBlock != nil {
				features[FeatureIngressPeerIPBlock] = true
				if len(peer.IPBlock.Except) == 0 {
					features[FeatureIngressPeerIPBlockEmptyExcept] = true
				} else {
					features[FeatureIngressPeerIPBlockNonemptyExcept] = true
				}
			} else {
				if peer.PodSelector != nil {
					sel := *peer.PodSelector
					if len(sel.MatchLabels) == 0 && len(sel.MatchExpressions) == 0 {
						features[FeatureIngressPeerPodSelectorEmpty] = true
					}
					if len(sel.MatchLabels) > 0 {
						features[FeatureIngressPeerPodSelectorMatchLabels] = true
					}
					if len(sel.MatchExpressions) > 0 {
						features[FeatureIngressPeerPodSelectorMatchExpressions] = true
					}
				} else {
					features[FeatureIngressPeerPodSelectorNil] = true
				}
				if peer.NamespaceSelector != nil {
					sel := peer.NamespaceSelector
					if len(sel.MatchLabels) == 0 && len(sel.MatchExpressions) == 0 {
						features[FeatureIngressPeerNamespaceSelectorEmpty] = true
					}
					if len(sel.MatchLabels) > 0 {
						features[FeatureIngressPeerNamespaceSelectorMatchLabels] = true
					}
					if len(sel.MatchExpressions) > 0 {
						features[FeatureIngressPeerNamespaceSelectorMatchExpressions] = true
					}
				} else {
					features[FeatureIngressPeerNamespaceSelectorNil] = true
				}
			}
		}
	}
	return features
}

func peerEgressFeatures(peers []networkingv1.NetworkPolicyPeer) map[Feature]bool {
	features := map[Feature]bool{}
	if len(peers) == 0 {
		features[FeatureEgressEmptyPeerSlice] = true
	} else {
		for _, peer := range peers {
			if peer.IPBlock != nil {
				features[FeatureEgressPeerIPBlock] = true
				if len(peer.IPBlock.Except) == 0 {
					features[FeatureEgressPeerIPBlockEmptyExcept] = true
				} else {
					features[FeatureEgressPeerIPBlockNonemptyExcept] = true
				}
			} else {
				if peer.PodSelector != nil {
					sel := *peer.PodSelector
					if len(sel.MatchLabels) == 0 && len(sel.MatchExpressions) == 0 {
						features[FeatureEgressPeerPodSelectorEmpty] = true
					}
					if len(sel.MatchLabels) > 0 {
						features[FeatureEgressPeerPodSelectorMatchLabels] = true
					}
					if len(sel.MatchExpressions) > 0 {
						features[FeatureEgressPeerPodSelectorMatchExpressions] = true
					}
				} else {
					features[FeatureEgressPeerPodSelectorNil] = true
				}
				if peer.NamespaceSelector != nil {
					sel := peer.NamespaceSelector
					if len(sel.MatchLabels) == 0 && len(sel.MatchExpressions) == 0 {
						features[FeatureEgressPeerNamespaceSelectorEmpty] = true
					}
					if len(sel.MatchLabels) > 0 {
						features[FeatureEgressPeerNamespaceSelectorMatchLabels] = true
					}
					if len(sel.MatchExpressions) > 0 {
						features[FeatureEgressPeerNamespaceSelectorMatchExpressions] = true
					}
				} else {
					features[FeatureEgressPeerNamespaceSelectorNil] = true
				}
			}
		}
	}
	return features
}

func portIngressFeatures(npPorts []networkingv1.NetworkPolicyPort) map[Feature]bool {
	features := map[Feature]bool{}
	if len(npPorts) == 0 {
		features[FeatureIngressEmptyPortSlice] = true
	} else {
		for _, npPort := range npPorts {
			if npPort.Port == nil {
				features[FeatureIngressNilPort] = true
			} else {
				switch (*npPort.Port).Type {
				case intstr.Int:
					features[FeatureIngressNumberedPort] = true
				case intstr.String:
					features[FeatureIngressNamedPort] = true
				default:
					panic("invalid intstr value")
				}
			}
		}
	}
	return features
}

func portEgressFeatures(npPorts []networkingv1.NetworkPolicyPort) map[Feature]bool {
	features := map[Feature]bool{}
	if len(npPorts) == 0 {
		features[FeatureEgressEmptyPortSlice] = true
	} else {
		for _, npPort := range npPorts {
			if npPort.Port == nil {
				features[FeatureEgressNilPort] = true
			} else {
				switch (*npPort.Port).Type {
				case intstr.Int:
					features[FeatureEgressNumberedPort] = true
				case intstr.String:
					features[FeatureEgressNamedPort] = true
				default:
					panic("invalid intstr value")
				}
			}
		}
	}
	return features
}

func targetPodSelectorFeatures(sel metav1.LabelSelector) map[Feature]bool {
	features := map[Feature]bool{}
	if len(sel.MatchLabels) == 0 && len(sel.MatchExpressions) == 0 {
		features[FeatureTargetPodSelectorEmpty] = true
	}
	if len(sel.MatchLabels) > 0 {
		features[FeatureTargetPodSelectorMatchLabels] = true
	}
	if len(sel.MatchExpressions) > 0 {
		features[FeatureTargetPodSelectorMatchExpressions] = true
	}
	return features
}

func mergeSets(l, r map[Feature]bool) map[Feature]bool {
	merged := map[Feature]bool{}
	for k := range l {
		merged[k] = true
	}
	for k := range r {
		merged[k] = true
	}
	return merged
}
