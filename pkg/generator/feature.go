package generator

import (
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ActionFeature string

const (
	ActionFeatureCreatePolicy       ActionFeature = "ActionCreatePolicy"
	ActionFeatureDeletePolicy       ActionFeature = "ActionDeletePolicy"
	ActionFeatureReadPolicies       ActionFeature = "ActionReadPolicies"
	ActionFeatureSetPodLabels       ActionFeature = "ActionSetPodLabels"
	ActionFeatureSetNamespaceLabels ActionFeature = "ActionSetNamespaceLabels"
	ActionFeatureUpdatePolicy       ActionFeature = "ActionUpdatePolicy"
)

type PeerFeature string

const (
	PeerFeatureEmptyRuleSlice PeerFeature = "PeerEmptyRuleSlice"

	PeerFeaturePortSliceEmpty     PeerFeature = "PeerPortSliceEmpty"
	PeerFeaturePortSliceSize1     PeerFeature = "PeerPortSliceSize1"
	PeerFeaturePortSliceSize2Plus PeerFeature = "PeerPortSliceSize2Plus"
	PeerFeatureNumberedPort       PeerFeature = "PeerNumberedPort"
	PeerFeatureNamedPort          PeerFeature = "PeerNamedPort"
	PeerFeatureNilPort            PeerFeature = "PeerNilPort"
	PeerFeatureNilProtocol        PeerFeature = "PeerNilProtocol"
	PeerFeatureTCPProtocol        PeerFeature = "PeerTCPProtocol"
	PeerFeatureUDPProtocol        PeerFeature = "PeerUDPProtocol"
	PeerFeatureSCTPProtocol       PeerFeature = "PeerSCTPProtocol"

	PeerFeatureEmptyPeerSlice                    PeerFeature = "PeerEmptyPeerSlice"
	PeerFeatureIPBlock                           PeerFeature = "PeerIPBlock"
	PeerFeatureIPBlockEmptyExcept                PeerFeature = "PeerIPBlockEmptyExcept"
	PeerFeatureIPBlockNonemptyExcept             PeerFeature = "PeerIPBlockNonemptyExcept"
	PeerFeaturePodSelectorEmpty                  PeerFeature = "PeerPodSelectorEmpty"
	PeerFeaturePodSelectorMatchLabels            PeerFeature = "PeerPodSelectorMatchLabels"
	PeerFeaturePodSelectorMatchExpressions       PeerFeature = "PeerPodSelectorMatchExpressions"
	PeerFeaturePodSelectorNil                    PeerFeature = "PeerPodSelectorNil"
	PeerFeatureNamespaceSelectorEmpty            PeerFeature = "PeerNamespaceSelectorEmpty"
	PeerFeatureNamespaceSelectorMatchLabels      PeerFeature = "PeerNamespaceSelectorMatchLabels"
	PeerFeatureNamespaceSelectorMatchExpressions PeerFeature = "PeerNamespaceSelectorMatchExpressions"
	PeerFeatureNamespaceSelectorNil              PeerFeature = "PeerNamespaceSelectorNil"
)

type TargetFeature string

const (
	TargetFeatureNamespaceEmpty              TargetFeature = "TargetNamespaceEmpty"
	TargetFeaturePodSelectorEmpty            TargetFeature = "TargetPodSelectorEmpty"
	TargetFeaturePodSelectorMatchLabels      TargetFeature = "TargetPodSelectorMatchLabels"
	TargetFeaturePodSelectorMatchExpressions TargetFeature = "TargetPodSelectorMatchExpressions"
)

type PolicyFeature string

const (
	PolicyFeatureIngress          PolicyFeature = "PolicyIngress"
	PolicyFeatureEgress           PolicyFeature = "PolicyEgress"
	PolicyFeatureIngressAndEgress PolicyFeature = "PolicyIngressAndEgress"
)

/*
type ProbeFeature string

// TODO goal: track what's used in probes.  Problem: AllAvailable could include SCTP/UDP/TCP etc., which
//    this package doesn't know about.
const (
	ProbeFeatureAllAvailable ProbeFeature = "AllAvailable"
	ProbeFeatureNumberedPort ProbeFeature = "NumberedPort"
	ProbeFeatureNamedPort ProbeFeature = "NamedPort"
	ProbeFeatureTCP ProbeFeature = "TCP"
	ProbeFeatureUDP ProbeFeature = "UDP"
	ProbeFeatureSCTP ProbeFeature = "SCTP"
)
*/

type Features struct {
	Policy  map[PolicyFeature]bool
	Target  map[TargetFeature]bool
	Action  map[ActionFeature]bool
	Ingress map[PeerFeature]bool
	Egress  map[PeerFeature]bool
}

func (f *Features) Strings() []string {
	var strs []string
	for feature := range f.Policy {
		strs = append(strs, string(feature))
	}
	for feature := range f.Target {
		strs = append(strs, string(feature))
	}
	for feature := range f.Action {
		strs = append(strs, string(feature))
	}
	for feature := range f.Ingress {
		strs = append(strs, "Ingress"+string(feature))
	}
	for feature := range f.Egress {
		strs = append(strs, "Egress"+string(feature))
	}
	return strs
}

func (f *Features) Combine(other *Features) *Features {
	return &Features{
		Policy:  mergePolicyFeatureSets(f.Policy, other.Policy),
		Target:  mergeTargetFeatureSets(f.Target, other.Target),
		Action:  mergeActionFeatureSets(f.Action, other.Action),
		Ingress: mergePeerFeatureSets(f.Ingress, other.Ingress),
		Egress:  mergePeerFeatureSets(f.Egress, other.Egress),
	}
}

/*
// AllFeatures is a slice of all Features.  Make sure to keep it updated as new Features are created.
var AllFeatures = []string{
	string(ActionFeatureCreatePolicy),
	string(ActionFeatureDeletePolicy),
	string(ActionFeatureReadPolicies),
	string(ActionFeatureSetPodLabels),
	string(ActionFeatureSetNamespaceLabels),
	string(ActionFeatureUpdatePolicy),

	string(TargetFeatureNamespaceEmpty),
	string(TargetFeaturePodSelectorEmpty),
	string(TargetFeaturePodSelectorMatchLabels),
	string(TargetFeaturePodSelectorMatchExpressions),

	string(FeatureIngressEmptyPortSlice),
	string(FeatureIngressNumberedPort),
	string(FeatureIngressNamedPort),
	string(FeatureIngressNilPort),

	string(FeatureIngressEmptyPeerSlice),
	string(FeatureIngressPeerIPBlock),
	string(FeatureIngressPeerIPBlockEmptyExcept),
	string(FeatureIngressPeerIPBlockNonemptyExcept),
	string(FeatureIngressPeerPodSelectorEmpty),
	string(FeatureIngressPeerPodSelectorMatchLabels),
	string(FeatureIngressPeerPodSelectorMatchExpressions),
	string(FeatureIngressPeerPodSelectorNil),
	string(FeatureIngressPeerNamespaceSelectorEmpty),
	string(FeatureIngressPeerNamespaceSelectorMatchLabels),
	string(FeatureIngressPeerNamespaceSelectorMatchExpressions),
	string(FeatureIngressPeerNamespaceSelectorNil),

	string(FeatureIngressEmptySlice),

	string(PolicyFeatureIngress),
	string(PolicyFeatureEgress),
	string(PolicyFeatureIngressAndEgress),
}
*/

func GetFeaturesForPolicy(policy *networkingv1.NetworkPolicy) *Features {
	spec := policy.Spec
	targetFeatures := targetPodSelectorFeatures(spec.PodSelector)
	if policy.Namespace == "" {
		targetFeatures[TargetFeatureNamespaceEmpty] = true
	}

	ingress := map[PeerFeature]bool{}
	egress := map[PeerFeature]bool{}
	hasIngress, hasEegress := false, false
	for _, policyType := range spec.PolicyTypes {
		if policyType == networkingv1.PolicyTypeIngress {
			hasIngress = true
			ingress = mergePeerFeatureSets(ingress, ingressFeatures(spec.Ingress))
		} else if policyType == networkingv1.PolicyTypeEgress {
			hasEegress = true
			egress = mergePeerFeatureSets(egress, egressFeatures(spec.Egress))
		}
	}
	policyFeatures := map[PolicyFeature]bool{}
	if hasIngress {
		policyFeatures[PolicyFeatureIngress] = true
	}
	if hasEegress {
		policyFeatures[PolicyFeatureEgress] = true
	}
	if hasIngress && hasEegress {
		policyFeatures[PolicyFeatureIngressAndEgress] = true
	}
	return &Features{
		Policy:  policyFeatures,
		Target:  targetFeatures,
		Action:  nil,
		Ingress: ingress,
		Egress:  egress,
	}
}

func ingressFeatures(rules []networkingv1.NetworkPolicyIngressRule) map[PeerFeature]bool {
	features := map[PeerFeature]bool{}
	if len(rules) == 0 {
		features[PeerFeatureEmptyRuleSlice] = true
	} else {
		for _, rule := range rules {
			features = mergePeerFeatureSets(features, peerFeatures(rule.From))
			features = mergePeerFeatureSets(features, portFeatures(rule.Ports))
		}
	}
	return features
}

func egressFeatures(rules []networkingv1.NetworkPolicyEgressRule) map[PeerFeature]bool {
	features := map[PeerFeature]bool{}
	if len(rules) == 0 {
		features[PeerFeatureEmptyRuleSlice] = true
	} else {
		for _, rule := range rules {
			features = mergePeerFeatureSets(features, peerFeatures(rule.To))
			features = mergePeerFeatureSets(features, portFeatures(rule.Ports))
		}
	}
	return features
}

func peerFeatures(peers []networkingv1.NetworkPolicyPeer) map[PeerFeature]bool {
	features := map[PeerFeature]bool{}
	if len(peers) == 0 {
		features[PeerFeatureEmptyPeerSlice] = true
	} else {
		for _, peer := range peers {
			if peer.IPBlock != nil {
				features[PeerFeatureIPBlock] = true
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
	}
	return features
}

func portFeatures(npPorts []networkingv1.NetworkPolicyPort) map[PeerFeature]bool {
	features := map[PeerFeature]bool{}
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

func targetPodSelectorFeatures(sel metav1.LabelSelector) map[TargetFeature]bool {
	features := map[TargetFeature]bool{}
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

// oh, the copy/paste!

func mergeTargetFeatureSets(l, r map[TargetFeature]bool) map[TargetFeature]bool {
	merged := map[TargetFeature]bool{}
	for k := range l {
		merged[k] = true
	}
	for k := range r {
		merged[k] = true
	}
	return merged
}

func mergePeerFeatureSets(l, r map[PeerFeature]bool) map[PeerFeature]bool {
	merged := map[PeerFeature]bool{}
	for k := range l {
		merged[k] = true
	}
	for k := range r {
		merged[k] = true
	}
	return merged
}

func mergePolicyFeatureSets(l, r map[PolicyFeature]bool) map[PolicyFeature]bool {
	merged := map[PolicyFeature]bool{}
	for k := range l {
		merged[k] = true
	}
	for k := range r {
		merged[k] = true
	}
	return merged
}

func mergeActionFeatureSets(l, r map[ActionFeature]bool) map[ActionFeature]bool {
	merged := map[ActionFeature]bool{}
	for k := range l {
		merged[k] = true
	}
	for k := range r {
		merged[k] = true
	}
	return merged
}
