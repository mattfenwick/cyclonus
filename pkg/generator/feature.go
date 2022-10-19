package generator

import (
	v1 "k8s.io/api/core/v1"
	. "k8s.io/api/networking/v1"
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

	ActionFeatureCreateService = "action: create service"
	ActionFeatureDeleteService = "action: delete service"
)

const (
	PolicyFeatureIngress          = "policy with ingress"
	PolicyFeatureEgress           = "policy with egress"
	PolicyFeatureIngressAndEgress = "policy with both ingress and egress"
)

const (
	TargetFeatureSpecificNamespace           = "target: specific namespace"
	TargetFeatureNamespaceEmpty              = "target: empty namespace"
	TargetFeaturePodSelectorEmpty            = "target: empty pod selector"
	TargetFeaturePodSelectorMatchLabels      = "target: pod selector match labels"
	TargetFeaturePodSelectorMatchExpressions = "target: pod selector match expression"
)

const (
	RuleFeatureAllPeersAllPortsAllProtocols = "all peers on all ports/protocols"

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

type NetpolTraverser struct {
	policy       func(*Netpol, map[string]bool)
	target       func(*NetpolTarget, map[string]bool)
	ingress      func(bool, *NetpolPeers, map[string]bool)
	ingressRule  func(bool, *Rule, map[string]bool)
	ingressPeers func(bool, []NetworkPolicyPeer, map[string]bool)
	ingressPeer  func(bool, NetworkPolicyPeer, map[string]bool)
	ingressPorts func(bool, []NetworkPolicyPort, map[string]bool)
	ingressPort  func(bool, NetworkPolicyPort, map[string]bool)
	egress       func(bool, *NetpolPeers, map[string]bool)
	egressRule   func(bool, *Rule, map[string]bool)
	egressPeers  func(bool, []NetworkPolicyPeer, map[string]bool)
	egressPeer   func(bool, NetworkPolicyPeer, map[string]bool)
	egressPorts  func(bool, []NetworkPolicyPort, map[string]bool)
	egressPort   func(bool, NetworkPolicyPort, map[string]bool)
}

var (
	GeneralNetpolTraverser = &NetpolTraverser{
		policy: DefaultPolicyFeatures,
		target: DefaultTargetFeatures,
	}

	IngressNetpolTraverser = &NetpolTraverser{
		ingress:      DefaultIngressOrEgressFeatures,
		ingressRule:  DefaultRuleFeature,
		ingressPeers: DefaultPeerFeatures,
		ingressPeer:  DefaultSinglePeerFeature,
		ingressPorts: DefaultPortFeatures,
		ingressPort:  DefaultSinglePortFeature,
	}

	EgressNetpolTraverser = &NetpolTraverser{
		egress:      DefaultIngressOrEgressFeatures,
		egressRule:  DefaultRuleFeature,
		egressPeers: DefaultPeerFeatures,
		egressPeer:  DefaultSinglePeerFeature,
		egressPorts: DefaultPortFeatures,
		egressPort:  DefaultSinglePortFeature,
	}
)

func (n *NetpolTraverser) Traverse(policy *Netpol) map[string]bool {
	features := map[string]bool{}
	if n.policy != nil {
		n.policy(policy, features)
	}
	if n.target != nil {
		n.target(policy.Target, features)
	}
	if policy.Ingress != nil {
		traverseIngressEgress(true, policy.Ingress, features, n.ingress, n.ingressRule, n.ingressPeers, n.ingressPeer, n.ingressPorts, n.ingressPort)
	}
	if policy.Egress != nil {
		traverseIngressEgress(false, policy.Egress, features, n.egress, n.egressRule, n.egressPeers, n.egressPeer, n.egressPorts, n.egressPort)
	}
	return features
}

func traverseIngressEgress(
	isIngress bool,
	netpolPeers *NetpolPeers,
	features map[string]bool,
	ingressEgress func(bool, *NetpolPeers, map[string]bool),
	rule func(bool, *Rule, map[string]bool),
	peers func(bool, []NetworkPolicyPeer, map[string]bool),
	peer func(bool, NetworkPolicyPeer, map[string]bool),
	ports func(bool, []NetworkPolicyPort, map[string]bool),
	port func(bool, NetworkPolicyPort, map[string]bool),
) {
	if ingressEgress != nil {
		ingressEgress(isIngress, netpolPeers, features)
	}
	for _, peerRule := range netpolPeers.Rules {
		if rule != nil {
			rule(isIngress, peerRule, features)
		}
		if peers != nil {
			peers(isIngress, peerRule.Peers, features)
		}
		for _, rulePeer := range peerRule.Peers {
			if peer != nil {
				peer(isIngress, rulePeer, features)
			}
		}
		if ports != nil {
			ports(isIngress, peerRule.Ports, features)
		}
		for _, rulePort := range peerRule.Ports {
			if port != nil {
				port(isIngress, rulePort, features)
			}
		}
	}
}

func DefaultPolicyFeatures(policy *Netpol, features map[string]bool) {
	hasIngress := len(policy.Ingress.Rules) > 0
	if hasIngress {
		features[PolicyFeatureIngress] = true
	}

	hasEgress := len(policy.Egress.Rules) > 0
	if hasEgress {
		features[PolicyFeatureEgress] = true
	}

	if hasIngress && hasEgress {
		features[PolicyFeatureIngressAndEgress] = true
	}
}

func DefaultTargetFeatures(target *NetpolTarget, features map[string]bool) {
	if target.Namespace == "" {
		features[TargetFeatureNamespaceEmpty] = true
	} else {
		features[TargetFeatureSpecificNamespace] = true
	}

	selector := target.PodSelector
	if len(selector.MatchLabels) == 0 && len(selector.MatchExpressions) == 0 {
		features[TargetFeaturePodSelectorEmpty] = true
	}
	if len(selector.MatchLabels) > 0 {
		features[TargetFeaturePodSelectorMatchLabels] = true
	}
	if len(selector.MatchExpressions) > 0 {
		features[TargetFeaturePodSelectorMatchExpressions] = true
	}
}

func DefaultIngressOrEgressFeatures(isIngress bool, peers *NetpolPeers, features map[string]bool) {
	if peers != nil {
		switch len(peers.Rules) {
		case 0:
			features[RuleFeatureSliceEmpty] = true
		case 1:
			features[RuleFeatureSliceSize1] = true
		default:
			features[RuleFeatureSliceSize2Plus] = true
		}
	}
}

func DefaultRuleFeature(isIngress bool, rule *Rule, features map[string]bool) {
	if len(rule.Ports) == 0 && len(rule.Peers) == 0 {
		features[RuleFeatureAllPeersAllPortsAllProtocols] = true
	}
}

func DefaultPeerFeatures(isIngress bool, peers []NetworkPolicyPeer, features map[string]bool) {
	switch len(peers) {
	case 0:
		features[PeerFeaturePeerSliceEmpty] = true
	case 1:
		features[PeerFeaturePeerSliceSize1] = true
	default:
		features[PeerFeaturePeerSliceSize2Plus] = true
	}
}

func DefaultSinglePeerFeature(isIngress bool, peer NetworkPolicyPeer, features map[string]bool) {
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

func DefaultPortFeatures(isIngress bool, npPorts []NetworkPolicyPort, features map[string]bool) {
	switch len(npPorts) {
	case 0:
		features[PeerFeaturePortSliceEmpty] = true
	case 1:
		features[PeerFeaturePortSliceSize1] = true
	default:
		features[PeerFeaturePortSliceSize2Plus] = true
	}
}

func DefaultSinglePortFeature(isIngress bool, npPort NetworkPolicyPort, features map[string]bool) {
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
