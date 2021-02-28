package generator

import (
	"fmt"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	AllowDNSRule = &Rule{
		Ports: []NetworkPolicyPort{
			{
				Protocol: &udp,
				Port:     &port53,
			},
		},
	}

	AllowDNSPeers = &NetpolPeers{
		Rules: []*Rule{AllowDNSRule},
	}
)

func AllowDNSPolicy(source *NetpolTarget) *Netpol {
	return &Netpol{
		Name:   "allow-dns",
		Target: source,
		Egress: AllowDNSPeers,
	}
}

var (
	emptyPort = NetworkPolicyPort{
		Protocol: nil,
		Port:     nil,
	}
	sctpOnAnyPort = NetworkPolicyPort{
		Protocol: &sctp,
		Port:     nil,
	}
	implicitTCPOnPort80 = NetworkPolicyPort{
		Protocol: nil,
		Port:     &port80,
	}
	explicitUDPOnPort80 = NetworkPolicyPort{
		Protocol: &udp,
		Port:     &port80,
	}
	namedPort81TPCP = NetworkPolicyPort{
		Protocol: &tcp,
		Port:     &portServe81TCP,
	}
)

var (
	emptySliceOfPorts = []NetworkPolicyPort{}
)

func DefaultPorts() []NetworkPolicyPort {
	return []NetworkPolicyPort{
		emptyPort,
		sctpOnAnyPort,
		implicitTCPOnPort80,
		explicitUDPOnPort80,
	}
}

var (
	nilSelector                   *metav1.LabelSelector
	emptySelector                 = &metav1.LabelSelector{}
	podAMatchLabelsSelector       = &metav1.LabelSelector{MatchLabels: map[string]string{"pod": "a"}}
	podCMatchLabelsSelector       = &metav1.LabelSelector{MatchLabels: map[string]string{"pod": "c"}}
	podABMatchExpressionsSelector = &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "pod",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"a", "b"},
			},
		},
	}
	podBCMatchExpressionsSelector = &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "pod",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"b", "c"},
			},
		},
	}

	nsXMatchLabelsSelector       = &metav1.LabelSelector{MatchLabels: map[string]string{"ns": "x"}}
	nsXYMatchExpressionsSelector = &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "ns",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"x", "y"},
			},
		},
	}
	nsYZMatchExpressionsSelector = &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "ns",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"y", "z"},
			},
		},
	}
)

func DefaultIPBlockPeers(podIP string) []NetworkPolicyPeer {
	cidr24 := fmt.Sprintf("%s/24", podIP)
	//cidr26 := fmt.Sprintf("%s/26", podIP)
	cidr28 := fmt.Sprintf("%s/28", podIP)
	//cidr30 := fmt.Sprintf("%s/30", podIP)
	return []NetworkPolicyPeer{
		{
			IPBlock: &IPBlock{
				CIDR:   cidr24,
				Except: nil,
			},
		},
		{
			IPBlock: &IPBlock{
				CIDR:   cidr24,
				Except: []string{cidr28},
			},
		},
	}
}

func DefaultPodPeers() []NetworkPolicyPeer {
	var peers []NetworkPolicyPeer
	for _, nsSel := range []*metav1.LabelSelector{nilSelector, emptySelector, nsXMatchLabelsSelector} {
		for _, podSel := range []*metav1.LabelSelector{nilSelector, emptySelector, podCMatchLabelsSelector} {
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
	return peers
}

func DefaultPeers(podIP string) []NetworkPolicyPeer {
	return append(DefaultPodPeers(), DefaultIPBlockPeers(podIP)...)
}

var (
	emptySliceOfPeers = []NetworkPolicyPeer{}
)

var (
	emptySliceOfRules = []*Rule{}
)

func DefaultTargets() []metav1.LabelSelector {
	return []metav1.LabelSelector{
		*emptySelector,
		*podAMatchLabelsSelector,
		*podABMatchExpressionsSelector,
	}
}

func DefaultNamespaces() []string {
	return []string{
		"x",
		"y",
		"z",
	}
}

var (
	TypicalNamespace = "x"
	TypicalTarget    = metav1.LabelSelector{
		MatchLabels:      map[string]string{"pod": "a"},
		MatchExpressions: nil,
	}
	TypicalPorts = []NetworkPolicyPort{{Protocol: &tcp, Port: &port80}}
	TypicalPeers = []NetworkPolicyPeer{
		{
			PodSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"pod": "b"},
			},
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"ns": "y"},
			},
		},
	}
)
