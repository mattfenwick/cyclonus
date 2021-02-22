package generator

import (
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ActionFeatureCreatePolicy = "action: create policy"
	ActionFeatureUpdatePolicy = "action: update policy"
	ActionFeatureDeletePolicy = "action: delete policy"

	ActionFeatureCreateNamespace    = "action: create namespace"
	ActionFeatureSetNamespaceLabels = "action: set namespace labels"
	ActionFeatureDeleteNamespace    = "action: delete namespace"

	ActionFeatureReadPolicies = "action: read policies"

	ActionFeatureCreatePod    = "action: create pod"
	ActionFeatureSetPodLabels = "action: set pod labels"
	ActionFeatureDeletePod    = "action: delete pod"
)

const (
	RuleFeatureSliceEmpty     = "0 rules"
	RuleFeatureSliceSize1     = "1 rule"
	RuleFeatureSliceSize2Plus = "2+ rules"

	PeerFeaturePortSliceEmpty     = "0 port/protocols"
	PeerFeaturePortSliceSize1     = "1 port/protocol"
	PeerFeaturePortSliceSize2Plus = "2+ port/protocols"
	PeerFeatureNumberedPort       = "numbered port"
	PeerFeatureNamedPort          = "named port"
	PeerFeatureNilPort            = "nil port"
	PeerFeatureNilProtocol        = "nil protocol"
	PeerFeatureTCPProtocol        = "policy on TCP"
	PeerFeatureUDPProtocol        = "policy on UDP"
	PeerFeatureSCTPProtocol       = "policy on SCTP"

	PeerFeaturePeerSliceEmpty                    = "0 peers"
	PeerFeaturePeerSliceSize1                    = "1 peer"
	PeerFeaturePeerSliceSize2Plus                = "2+ peers"
	PeerFeatureIPBlockEmptyExcept                = "IPBlock (no except)"
	PeerFeatureIPBlockNonemptyExcept             = "IPBlock with except"
	PeerFeaturePodSelectorNil                    = "peer pod selector nil"
	PeerFeaturePodSelectorEmpty                  = "peer pod selector empty"
	PeerFeaturePodSelectorMatchLabels            = "peer pod selector match labels"
	PeerFeaturePodSelectorMatchExpressions       = "peer pod selector match expression"
	PeerFeatureNamespaceSelectorNil              = "peer namespace selector nil"
	PeerFeatureNamespaceSelectorEmpty            = "peer namespace selector empty"
	PeerFeatureNamespaceSelectorMatchLabels      = "peer namespace selector match labels"
	PeerFeatureNamespaceSelectorMatchExpressions = "peer namespace selector match expression"
)

const (
	TargetFeatureNamespaceEmpty              = "target: empty namespace"
	TargetFeaturePodSelectorEmpty            = "target: empty pod selector"
	TargetFeaturePodSelectorMatchLabels      = "target: pod selector match labels"
	TargetFeaturePodSelectorMatchExpressions = "target: pod selector match expression"
)

const (
	PolicyFeatureIngress          = "policy with ingress"
	PolicyFeatureEgress           = "policy with egress"
	PolicyFeatureIngressAndEgress = "policy with both ingress and egress"
)

func GetFeaturesForPolicy(policy *networkingv1.NetworkPolicy) map[string]bool {
	spec := policy.Spec
	features := targetPodSelectorFeatures(spec.PodSelector)
	if policy.Namespace == "" {
		features[TargetFeatureNamespaceEmpty] = true
	}

	ingress := map[string]bool{}
	egress := map[string]bool{}
	hasIngress, hasEegress := false, false
	for _, policyType := range spec.PolicyTypes {
		if policyType == networkingv1.PolicyTypeIngress {
			hasIngress = true
			ingress = mergeSets(ingress, addDirectionalityToKeys(true, ingressFeatures(spec.Ingress)))
		} else if policyType == networkingv1.PolicyTypeEgress {
			hasEegress = true
			egress = mergeSets(egress, addDirectionalityToKeys(false, egressFeatures(spec.Egress)))
		}
	}
	if hasIngress {
		features[PolicyFeatureIngress] = true
	}
	if hasEegress {
		features[PolicyFeatureEgress] = true
	}
	if hasIngress && hasEegress {
		features[PolicyFeatureIngressAndEgress] = true
	}
	return features
}

func ingressFeatures(rules []networkingv1.NetworkPolicyIngressRule) map[string]bool {
	features := map[string]bool{}
	switch len(rules) {
	case 0:
		features[RuleFeatureSliceEmpty] = true
	case 1:
		features[RuleFeatureSliceSize1] = true
	default:
		features[RuleFeatureSliceSize2Plus] = true
	}
	for _, rule := range rules {
		features = mergeSets(features, peerFeatures(rule.From))
		features = mergeSets(features, portFeatures(rule.Ports))
	}
	return features
}

func egressFeatures(rules []networkingv1.NetworkPolicyEgressRule) map[string]bool {
	features := map[string]bool{}
	switch len(rules) {
	case 0:
		features[RuleFeatureSliceEmpty] = true
	case 1:
		features[RuleFeatureSliceSize1] = true
	default:
		features[RuleFeatureSliceSize2Plus] = true
	}
	for _, rule := range rules {
		features = mergeSets(features, peerFeatures(rule.To))
		features = mergeSets(features, portFeatures(rule.Ports))
	}
	return features
}

func peerFeatures(peers []networkingv1.NetworkPolicyPeer) map[string]bool {
	features := map[string]bool{}
	switch len(peers) {
	case 0:
		features[PeerFeaturePeerSliceEmpty] = true
	case 1:
		features[PeerFeaturePeerSliceSize1] = true
	default:
		features[PeerFeaturePeerSliceSize2Plus] = true
	}
	for _, peer := range peers {
		if peer.IPBlock != nil {
			if len(peer.IPBlock.Except) == 0 {
				features[PeerFeatureIPBlockEmptyExcept] = true
			} else {
				features[PeerFeatureIPBlockNonemptyExcept] = true
			}
		} else {
			if peer.PodSelector != nil {
				sel := *peer.PodSelector
				if len(sel.MatchLabels) == 0 && len(sel.MatchExpressions) == 0 {
					features[PeerFeaturePodSelectorEmpty] = true
				}
				if len(sel.MatchLabels) > 0 {
					features[PeerFeaturePodSelectorMatchLabels] = true
				}
				if len(sel.MatchExpressions) > 0 {
					features[PeerFeaturePodSelectorMatchExpressions] = true
				}
			} else {
				features[PeerFeaturePodSelectorNil] = true
			}
			if peer.NamespaceSelector != nil {
				sel := peer.NamespaceSelector
				if len(sel.MatchLabels) == 0 && len(sel.MatchExpressions) == 0 {
					features[PeerFeatureNamespaceSelectorEmpty] = true
				}
				if len(sel.MatchLabels) > 0 {
					features[PeerFeatureNamespaceSelectorMatchLabels] = true
				}
				if len(sel.MatchExpressions) > 0 {
					features[PeerFeatureNamespaceSelectorMatchExpressions] = true
				}
			} else {
				features[PeerFeatureNamespaceSelectorNil] = true
			}
		}
	}
	return features
}

func portFeatures(npPorts []networkingv1.NetworkPolicyPort) map[string]bool {
	features := map[string]bool{}
	switch len(npPorts) {
	case 0:
		features[PeerFeaturePortSliceEmpty] = true
	case 1:
		features[PeerFeaturePortSliceSize1] = true
	default:
		features[PeerFeaturePortSliceSize2Plus] = true
	}
	for _, npPort := range npPorts {
		if npPort.Port == nil {
			features[PeerFeatureNilPort] = true
		} else {
			switch (*npPort.Port).Type {
			case intstr.Int:
				features[PeerFeatureNumberedPort] = true
			case intstr.String:
				features[PeerFeatureNamedPort] = true
			default:
				panic("invalid intstr value")
			}
		}
		if npPort.Protocol == nil {
			features[PeerFeatureNilProtocol] = true
		} else {
			switch *npPort.Protocol {
			case v1.ProtocolTCP:
				features[PeerFeatureTCPProtocol] = true
			case v1.ProtocolUDP:
				features[PeerFeatureUDPProtocol] = true
			case v1.ProtocolSCTP:
				features[PeerFeatureSCTPProtocol] = true
			}
		}
	}
	return features
}

func targetPodSelectorFeatures(sel metav1.LabelSelector) map[string]bool {
	features := map[string]bool{}
	if len(sel.MatchLabels) == 0 && len(sel.MatchExpressions) == 0 {
		features[TargetFeaturePodSelectorEmpty] = true
	}
	if len(sel.MatchLabels) > 0 {
		features[TargetFeaturePodSelectorMatchLabels] = true
	}
	if len(sel.MatchExpressions) > 0 {
		features[TargetFeaturePodSelectorMatchExpressions] = true
	}
	return features
}

func addDirectionalityToKeys(isIngress bool, dict map[string]bool) map[string]bool {
	var prefix string
	if isIngress {
		prefix = "Ingress: "
	} else {
		prefix = "Egress: "
	}
	return addPrefixToKeys(prefix, dict)
}

func addPrefixToKeys(prefix string, dict map[string]bool) map[string]bool {
	out := map[string]bool{}
	for k := range dict {
		out[prefix+k] = true
	}
	return out
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
