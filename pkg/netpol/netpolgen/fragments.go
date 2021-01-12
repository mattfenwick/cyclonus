package netpolgen

import (
	v1 "k8s.io/api/core/v1"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

/*
```
Test cases:

1 policy with ingress:
 - empty ingress
 - ingress with 1 rule
   - empty
   - 1 port
     - empty
     - protocol
     - port
     - port + protocol
   - 2 ports
   - 1 from
     - 8 combos: (nil + nil => might mean ipblock must be non-nil)
       - pod sel: nil, empty, non-empty
       - ns sel: nil, empty, non-empty
     - ipblock
       - no except
       - yes except
   - 2 froms
     - 1 pod/ns, 1 ipblock
     - 2 pod/ns
     - 2 ipblocks
   - 1 port, 1 from
   - 2 ports, 2 froms
 - ingress with 2 rules
 - ingress with 3 rules
2 policies with ingress
1 policy with egress
2 policies with egress
1 policy with both ingress and egress
2 policies with both ingress and egress
```
*/

var (
	sctp = v1.ProtocolSCTP
	tcp  = v1.ProtocolTCP
	udp  = v1.ProtocolUDP

	port99 = intstr.FromInt(80)
)

var (
	emptyPort = NetworkPolicyPort{
		Protocol: nil,
		Port:     nil,
	}
	protocolPort = NetworkPolicyPort{
		Protocol: &sctp,
		Port:     nil,
	}
	portPort = NetworkPolicyPort{
		Protocol: nil,
		Port:     &port99,
	}
	portProtocolPort = NetworkPolicyPort{
		Protocol: &udp,
		Port:     &port99,
	}
)

var (
	noPorts = []NetworkPolicyPort{}
)

func DefaultPorts() []NetworkPolicyPort {
	return []NetworkPolicyPort{
		emptyPort,
		protocolPort,
		portPort,
		portProtocolPort,
	}
}

var (
	ipNoExcept = IPBlock{
		CIDR:   "1.2.3.4/24",
		Except: nil,
	}
	ipWithExcept = IPBlock{
		CIDR:   "1.2.3.4/24",
		Except: []string{"1.2.3.8", "1.2.3.10"},
	}

	ipPeer1 = NetworkPolicyPeer{IPBlock: &ipNoExcept}
	ipPeer2 = NetworkPolicyPeer{IPBlock: &ipWithExcept}
)

var (
	nilPodSelector         *metav1.LabelSelector
	emptyPodSelector       = &metav1.LabelSelector{}
	matchLabelsPodSelector = &metav1.LabelSelector{
		MatchLabels:      map[string]string{"pod": "a"},
		MatchExpressions: nil,
	}

	nilNSSelector   *metav1.LabelSelector
	emptyNSSelector = &metav1.LabelSelector{}
	matchNSSelector = &metav1.LabelSelector{
		MatchLabels:      map[string]string{"ns": "x"},
		MatchExpressions: nil,
	}
)

func DefaultPeers() []NetworkPolicyPeer {
	var peers []NetworkPolicyPeer
	for _, nsSel := range []*metav1.LabelSelector{nilNSSelector, emptyNSSelector, matchNSSelector} {
		for _, podSel := range []*metav1.LabelSelector{nilPodSelector, emptyPodSelector, matchLabelsPodSelector} {
			if nsSel == nil && podSel == nil {
				// skip this case -- this is where IPBlock needs to be non-nil
			} else {
				peers = append(peers, NetworkPolicyPeer{
					PodSelector:       podSel,
					NamespaceSelector: nsSel,
					IPBlock:           nil,
				})
			}
		}
	}
	return append(peers, ipPeer1, ipPeer2)
}

var (
	noPeers = []NetworkPolicyPeer{}
)

var (
	noIngressRules = []NetworkPolicyIngressRule{}
	noEgressRules  = []NetworkPolicyEgressRule{}
)

func DefaultTargets() []metav1.LabelSelector {
	return []metav1.LabelSelector{
		metav1.LabelSelector{},
		metav1.LabelSelector{
			MatchLabels: map[string]string{"pod": "a"},
		},
		metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "pod",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"a", "b"},
				},
			},
		},
	}
}

func DefaultNamespaces() []string {
	return []string{
		"x",
		"default",
		"y",
	}
}

var (
	illustrativeExample = &NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "abc",
			Namespace: "xyz",
		},
		Spec: NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels:      nil,
				MatchExpressions: nil,
			},
			Ingress: []NetworkPolicyIngressRule{
				{
					Ports: []NetworkPolicyPort{
						{
							Protocol: &sctp,
							Port:     &port99,
						},
					},
					From: []NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels:      nil,
								MatchExpressions: nil,
							},
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels:      nil,
								MatchExpressions: nil,
							},
							IPBlock: nil,
						},
						{
							PodSelector:       nil,
							NamespaceSelector: nil,
							IPBlock: &IPBlock{
								CIDR:   "1.2.3.4/24",
								Except: []string{"1.2.3.8"},
							},
						},
					},
				},
			},
			Egress: nil,
			PolicyTypes: []PolicyType{
				PolicyTypeIngress,
				PolicyTypeEgress,
			},
		},
	}
)
