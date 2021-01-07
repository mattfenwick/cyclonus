package matcher

import (
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildNetworkPolicy(policy *networkingv1.NetworkPolicy) *Policy {
	return BuildNetworkPolicies([]*networkingv1.NetworkPolicy{policy})
}

func BuildNetworkPolicies(netpols []*networkingv1.NetworkPolicy) *Policy {
	np := NewPolicy()
	for _, policy := range netpols {
		ingress, egress := BuildTarget(policy)
		if ingress != nil {
			np.AddTarget(true, ingress)
		}
		if egress != nil {
			np.AddTarget(false, egress)
		}
	}
	return np
}

func BuildTarget(netpol *networkingv1.NetworkPolicy) (*Target, *Target) {
	var ingress *Target
	var egress *Target
	for _, pType := range netpol.Spec.PolicyTypes {
		switch pType {
		case networkingv1.PolicyTypeIngress:
			ingress = &Target{
				Namespace:   netpol.Namespace,
				PodSelector: netpol.Spec.PodSelector,
				SourceRules: []*networkingv1.NetworkPolicy{netpol},
				Edge:        BuildIngressMatcher(netpol.Namespace, netpol.Spec.Ingress),
			}
		case networkingv1.PolicyTypeEgress:
			egress = &Target{
				Namespace:   netpol.Namespace,
				PodSelector: netpol.Spec.PodSelector,
				SourceRules: []*networkingv1.NetworkPolicy{netpol},
				Edge:        BuildEgressMatcher(netpol.Namespace, netpol.Spec.Egress),
			}
		}
	}
	return ingress, egress
}

func BuildIngressMatcher(policyNamespace string, ingresses []networkingv1.NetworkPolicyIngressRule) EdgeMatcher {
	if len(ingresses) == 0 {
		return &NoneEdgeMatcher{}
	}
	var sdaps []*PeerPortMatcher
	for _, ingress := range ingresses {
		sdaps = append(sdaps, BuildPeerPortMatchers(policyNamespace, ingress.Ports, ingress.From)...)
	}
	return &EdgePeerPortMatcher{Matchers: sdaps}
}

func BuildEgressMatcher(policyNamespace string, egresses []networkingv1.NetworkPolicyEgressRule) EdgeMatcher {
	if len(egresses) == 0 {
		return &NoneEdgeMatcher{}
	}
	var sdaps []*PeerPortMatcher
	for _, egress := range egresses {
		sdaps = append(sdaps, BuildPeerPortMatchers(policyNamespace, egress.Ports, egress.To)...)
	}
	return &EdgePeerPortMatcher{Matchers: sdaps}
}

func BuildPeerPortMatchers(policyNamespace string, npPorts []networkingv1.NetworkPolicyPort, peers []networkingv1.NetworkPolicyPeer) []*PeerPortMatcher {
	// 1. build ports
	ports := BuildPortMatchers(npPorts)
	// 2. build SourceDests
	sds := BuildPeerMatchers(policyNamespace, peers)
	// 3. build the cartesian product of ports and SourceDests
	var sdaps []*PeerPortMatcher
	for _, port := range ports {
		for _, sd := range sds {
			sdaps = append(sdaps, &PeerPortMatcher{
				Peer: sd,
				Port: port,
			})
		}
	}
	return sdaps
}

func BuildPortMatchers(npPorts []networkingv1.NetworkPolicyPort) []PortMatcher {
	var ports []PortMatcher
	if len(npPorts) == 0 {
		ports = append(ports, &AllPortsAllProtocolsMatcher{})
	} else {
		for _, p := range npPorts {
			ports = append(ports, BuildPortMatcher(p))
		}
	}
	return ports
}

func BuildPortMatcher(p networkingv1.NetworkPolicyPort) PortMatcher {
	protocol := v1.ProtocolTCP
	if p.Protocol != nil {
		protocol = *p.Protocol
	}
	if p.Port == nil {
		return &AllPortsOnProtocolMatcher{Protocol: protocol}
	}
	return &ExactPortProtocolMatcher{Port: *p.Port, Protocol: protocol}
}

func BuildPeerMatchers(policyNamespace string, peers []networkingv1.NetworkPolicyPeer) []PeerMatcher {
	var sds []PeerMatcher
	if len(peers) == 0 {
		sds = append(sds, &AnywherePeerMatcher{})
	} else {
		for _, from := range peers {
			sds = append(sds, BuildPeerMatcher(policyNamespace, from))
		}
	}
	return sds
}

func isLabelSelectorEmpty(l metav1.LabelSelector) bool {
	return len(l.MatchLabels) == 0 && len(l.MatchExpressions) == 0
}

func BuildPeerMatcher(policyNamespace string, peer networkingv1.NetworkPolicyPeer) PeerMatcher {
	if peer.IPBlock != nil {
		return &IPBlockPeerMatcher{peer.IPBlock}
	}
	podSel := peer.PodSelector
	nsSel := peer.NamespaceSelector
	if podSel == nil || isLabelSelectorEmpty(*podSel) {
		if nsSel == nil {
			return &AllPodsInPolicyNamespacePeerMatcher{Namespace: policyNamespace}
		} else if isLabelSelectorEmpty(*nsSel) {
			return &AllPodsAllNamespacesPeerMatcher{}
		} else {
			// nsSel has some stuff
			return &AllPodsInMatchingNamespacesPeerMatcher{NamespaceSelector: *nsSel}
		}
	} else {
		// podSel has some stuff
		if nsSel == nil {
			return &MatchingPodsInPolicyNamespacePeerMatcher{
				PodSelector: *podSel,
				Namespace:   policyNamespace,
			}
		} else if isLabelSelectorEmpty(*nsSel) {
			return &MatchingPodsInAllNamespacesPeerMatcher{PodSelector: *podSel}
		} else {
			// nsSel has some stuff
			return &MatchingPodsInMatchingNamespacesPeerMatcher{
				PodSelector:       *podSel,
				NamespaceSelector: *nsSel,
			}
		}
	}
}
