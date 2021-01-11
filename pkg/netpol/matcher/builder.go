package matcher

import (
	"github.com/pkg/errors"
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
				Peer:        BuildIngressMatcher(netpol.Namespace, netpol.Spec.Ingress),
			}
		case networkingv1.PolicyTypeEgress:
			egress = &Target{
				Namespace:   netpol.Namespace,
				PodSelector: netpol.Spec.PodSelector,
				SourceRules: []*networkingv1.NetworkPolicy{netpol},
				Peer:        BuildEgressMatcher(netpol.Namespace, netpol.Spec.Egress),
			}
		}
	}
	return ingress, egress
}

func BuildIngressMatcher(policyNamespace string, ingresses []networkingv1.NetworkPolicyIngressRule) PeerMatcher {
	var matcher PeerMatcher = &NonePeerMatcher{}
	for _, ingress := range ingresses {
		matcher = CombinePeerMatchers(matcher, BuildPeerMatcher(policyNamespace, ingress.Ports, ingress.From))
	}
	return matcher
}

func BuildEgressMatcher(policyNamespace string, egresses []networkingv1.NetworkPolicyEgressRule) PeerMatcher {
	var matcher PeerMatcher = &NonePeerMatcher{}
	for _, egress := range egresses {
		matcher = CombinePeerMatchers(matcher, BuildPeerMatcher(policyNamespace, egress.Ports, egress.To))
	}
	return matcher
}

func BuildPeerMatcher(policyNamespace string, npPorts []networkingv1.NetworkPolicyPort, peers []networkingv1.NetworkPolicyPeer) PeerMatcher {
	// 1. build port matcher
	port := BuildPortMatcher(npPorts)
	// 2. build Peers
	if len(peers) == 0 {
		return &AllPeerMatcher{}
	} else {
		matcher := &SpecificPeerMatcher{
			IP:       map[string]*IPBlockMatcher{},
			Internal: &NoneInternalMatcher{},
		}
		for _, from := range peers {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(policyNamespace, from)
			// invalid netpol guards
			if ip == nil && ns == nil && pod == nil {
				panic(errors.Errorf("invalid NetworkPolicyPeer: all of IPBlock, NamespaceSelector, and PodSelector are nil"))
			}
			if ip != nil && (ns != nil || pod != nil) {
				panic(errors.Errorf("invalid NetworkPolicyPeer: if NamespaceSelector or PodSelector is non-nil, IPBlock must be nil"))
			}
			// process a valid netpol
			if ip != nil {
				ip.Port = port
				matcher.AddIPMatcher(ip)
			} else {
				// special case: if all ports, namespaces, and pods are allowed
				switch port.(type) {
				case *AllPortMatcher:
					switch ns.(type) {
					case *AllNamespaceMatcher:
						switch pod.(type) {
						case *AllPodMatcher:
							matcher.Internal = &AllInternalMatcher{}
						}
					}
				}
				// it's okay to continue processing additional matchers after hitting the special case,
				//   since nothing can override an AllInternalMatcher
				internal := &SpecificInternalMatcher{NamespacePods: map[string]*NamespacePodMatcher{}}
				internal.Add(&NamespacePodMatcher{
					Namespace: ns,
					Pod:       pod,
					Port:      port,
				})
				matcher.Internal = CombineInternalMatchers(matcher.Internal, internal)
			}
		}
		return matcher
	}
}

func BuildIPBlockNamespacePodMatcher(policyNamespace string, peer networkingv1.NetworkPolicyPeer) (*IPBlockMatcher, NamespaceMatcher, PodMatcher) {
	if peer.IPBlock != nil {
		return &IPBlockMatcher{
			IPBlock: peer.IPBlock,
			Port:    nil, // remember to set this elsewhere!
		}, nil, nil
	}

	podSel := peer.PodSelector
	var podMatcher PodMatcher
	if podSel == nil || isLabelSelectorEmpty(*podSel) {
		podMatcher = &AllPodMatcher{}
	} else {
		podMatcher = &LabelSelectorPodMatcher{Selector: *podSel}
	}

	nsSel := peer.NamespaceSelector
	var nsMatcher NamespaceMatcher
	if nsSel == nil {
		nsMatcher = &ExactNamespaceMatcher{Namespace: policyNamespace}
	} else if isLabelSelectorEmpty(*nsSel) {
		nsMatcher = &AllNamespaceMatcher{}
	} else {
		nsMatcher = &LabelSelectorNamespaceMatcher{Selector: *nsSel}
	}

	return nil, nsMatcher, podMatcher
}

func BuildPortMatcher(npPorts []networkingv1.NetworkPolicyPort) PortMatcher {
	if len(npPorts) == 0 {
		return &AllPortMatcher{}
	} else {
		matcher := &SpecificPortMatcher{}
		for _, p := range npPorts {
			protocol := v1.ProtocolTCP
			if p.Protocol != nil {
				protocol = *p.Protocol
			}
			matcher.Ports = append(matcher.Ports, &PortProtocolMatcher{
				Port:     p.Port,
				Protocol: protocol,
			})
		}
		return matcher
	}
}

func isLabelSelectorEmpty(l metav1.LabelSelector) bool {
	return len(l.MatchLabels) == 0 && len(l.MatchExpressions) == 0
}
