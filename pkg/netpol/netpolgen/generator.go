package netpolgen

import (
	"fmt"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewDefaultGenerator() *Generator {
	return &Generator{
		Ports:            DefaultPorts(),
		Peers:            DefaultPeers(),
		Targets:          DefaultTargets(),
		Namespaces:       DefaultNamespaces(),
		TypicalPorts:     TypicalPorts,
		TypicalPeers:     TypicalPeers,
		TypicalTarget:    TypicalTarget,
		TypicalNamespace: TypicalNamespace,
	}
}

type Generator struct {
	// multidimensional generation
	Ports      []NetworkPolicyPort
	Peers      []NetworkPolicyPeer
	Targets    []metav1.LabelSelector
	Namespaces []string
	// unidimensional typicals
	TypicalPorts     []NetworkPolicyPort
	TypicalPeers     []NetworkPolicyPeer
	TypicalTarget    metav1.LabelSelector
	TypicalNamespace string
}

func (g *Generator) PortSlices() [][]NetworkPolicyPort {
	// length 0
	slices := [][]NetworkPolicyPort{nil, emptySliceOfPorts}
	// length 1
	for _, s := range g.Ports {
		slices = append(slices, []NetworkPolicyPort{s})
	}
	// length 2
	for i, s1 := range g.Ports {
		for j, s2 := range g.Ports {
			if i < j {
				slices = append(slices, []NetworkPolicyPort{s1, s2})
			}
		}
	}
	return slices
}

func (g *Generator) PeerSlices() [][]NetworkPolicyPeer {
	// 0 length
	slices := [][]NetworkPolicyPeer{nil, emptySliceOfPeers}
	// 1 length
	for _, p := range g.Peers {
		slices = append(slices, []NetworkPolicyPeer{p})
	}
	// 2 length
	for i, p1 := range g.Peers {
		for j, p2 := range g.Peers {
			if i < j {
				slices = append(slices, []NetworkPolicyPeer{p1, p2})
			}
		}
	}
	return slices
}

func (g *Generator) IngressRules() []NetworkPolicyIngressRule {
	var is []NetworkPolicyIngressRule
	for _, ports := range g.PortSlices() {
		for _, peers := range g.PeerSlices() {
			is = append(is, NetworkPolicyIngressRule{
				Ports: ports,
				From:  peers,
			})
		}
	}
	return is
}

func (g *Generator) IngressRuleSlices() [][]NetworkPolicyIngressRule {
	// 0 length
	slices := [][]NetworkPolicyIngressRule{nil, emptySliceOfIngressRules}
	// 1 length
	for _, ingress := range g.IngressRules() {
		slices = append(slices, []NetworkPolicyIngressRule{ingress})
	}
	// 2 length
	for i, ing1 := range g.IngressRules() {
		for j, ing2 := range g.IngressRules() {
			if i < j {
				slices = append(slices, []NetworkPolicyIngressRule{ing1, ing2})
			}
		}
	}
	return slices
}

func (g *Generator) EgressRules() []NetworkPolicyEgressRule {
	var is []NetworkPolicyEgressRule
	for _, ports := range g.PortSlices() {
		for _, peers := range g.PeerSlices() {
			is = append(is, NetworkPolicyEgressRule{
				Ports: ports,
				To:    peers,
			})
		}
	}
	return is
}

func (g *Generator) EgressRuleSlices() [][]NetworkPolicyEgressRule {
	// 0 length
	slices := [][]NetworkPolicyEgressRule{nil, emptySliceOfEgressRules}
	// 1 length
	for _, egress := range g.EgressRules() {
		slices = append(slices, []NetworkPolicyEgressRule{egress})
	}
	// 2 length
	for i, eng1 := range g.EgressRules() {
		for j, eng2 := range g.EgressRules() {
			if i < j {
				slices = append(slices, []NetworkPolicyEgressRule{eng1, eng2})
			}
		}
	}
	return slices
}

// single policies, unidimensional generation

func ingressExample(name string, ns string, target metav1.LabelSelector, ports []NetworkPolicyPort, peers []NetworkPolicyPeer) *NetworkPolicy {
	return &NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: NetworkPolicySpec{
			PodSelector: target,
			Ingress: []NetworkPolicyIngressRule{
				{
					Ports: ports,
					From:  peers,
				},
			},
			PolicyTypes: []PolicyType{
				PolicyTypeIngress,
			},
		},
	}
}

func (g *Generator) VaryNamespaceIngressPolicies() []*NetworkPolicy {
	var policies []*NetworkPolicy
	for i, ns := range g.Namespaces {
		name := fmt.Sprintf("vary-ns-ingress-%d", i)
		policies = append(policies, ingressExample(name, ns, g.TypicalTarget, g.TypicalPorts, g.TypicalPeers))
	}
	return policies
}

func (g *Generator) VaryTargetIngressPolicies() []*NetworkPolicy {
	var policies []*NetworkPolicy
	for i, target := range g.Targets {
		name := fmt.Sprintf("vary-target-ingress-%d", i)
		policies = append(policies, ingressExample(name, g.TypicalNamespace, target, g.TypicalPorts, g.TypicalPeers))
	}
	return policies
}

func (g *Generator) VaryPortsIngressPolicies() []*NetworkPolicy {
	var policies []*NetworkPolicy
	for i, ports := range g.PortSlices() {
		name := fmt.Sprintf("vary-ports-ingress-%d", i)
		policies = append(policies, ingressExample(name, g.TypicalNamespace, g.TypicalTarget, ports, g.TypicalPeers))
	}
	return policies
}

func (g *Generator) VaryPeersIngressPolicies() []*NetworkPolicy {
	var policies []*NetworkPolicy
	for i, peers := range g.PeerSlices() {
		name := fmt.Sprintf("vary-peers-ingress-%d", i)
		policies = append(policies, ingressExample(name, g.TypicalNamespace, g.TypicalTarget, g.TypicalPorts, peers))
	}
	return policies
}

func egressExample(name string, ns string, target metav1.LabelSelector, ports []NetworkPolicyPort, peers []NetworkPolicyPeer) *NetworkPolicy {
	return &NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: NetworkPolicySpec{
			PodSelector: target,
			Egress: []NetworkPolicyEgressRule{
				{
					Ports: ports,
					To:    peers,
				},
			},
			PolicyTypes: []PolicyType{
				PolicyTypeEgress,
			},
		},
	}
}

func (g *Generator) VaryNamespaceEgressPolicies(allowDNS bool) []*NetworkPolicy {
	var policies []*NetworkPolicy
	for i, ns := range g.Namespaces {
		name := fmt.Sprintf("vary-ns-egress-%d", i)
		policy := egressExample(name, ns, g.TypicalTarget, g.TypicalPorts, g.TypicalPeers)
		if allowDNS {
			policy.Spec.Egress = append(policy.Spec.Egress, AllowDNSEgressRule)
		}
		policies = append(policies, policy)
	}
	return policies
}

func (g *Generator) VaryTargetEgressPolicies(allowDNS bool) []*NetworkPolicy {
	var policies []*NetworkPolicy
	for i, target := range g.Targets {
		name := fmt.Sprintf("vary-target-egress-%d", i)
		policy := egressExample(name, g.TypicalNamespace, target, g.TypicalPorts, g.TypicalPeers)
		if allowDNS {
			policy.Spec.Egress = append(policy.Spec.Egress, AllowDNSEgressRule)
		}
		policies = append(policies, policy)
	}
	return policies
}

func (g *Generator) VaryPortsEgressPolicies(allowDNS bool) []*NetworkPolicy {
	var policies []*NetworkPolicy
	for i, ports := range g.PortSlices() {
		name := fmt.Sprintf("vary-ports-egress-%d", i)
		policy := egressExample(name, g.TypicalNamespace, g.TypicalTarget, ports, g.TypicalPeers)
		if allowDNS {
			policy.Spec.Egress = append(policy.Spec.Egress, AllowDNSEgressRule)
		}
		policies = append(policies, policy)
	}
	return policies
}

func (g *Generator) VaryPeersEgressPolicies(allowDNS bool) []*NetworkPolicy {
	var policies []*NetworkPolicy
	for i, peers := range g.PeerSlices() {
		name := fmt.Sprintf("vary-peers-egress-%d", i)
		policy := egressExample(name, g.TypicalNamespace, g.TypicalTarget, g.TypicalPorts, peers)
		if allowDNS {
			policy.Spec.Egress = append(policy.Spec.Egress, AllowDNSEgressRule)
		}
		policies = append(policies, policy)
	}
	return policies
}

// single policies, multidimensional generation

func (g *Generator) IngressPolicies() []*NetworkPolicy {
	var policies []*NetworkPolicy
	i := 0
	for _, ingress := range g.IngressRuleSlices() {
		for _, target := range g.Targets {
			for _, ns := range g.Namespaces {
				policies = append(policies, &NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("policy-%d", i),
						Namespace: ns,
					},
					Spec: NetworkPolicySpec{
						PodSelector: target,
						Ingress:     ingress,
						PolicyTypes: []PolicyType{PolicyTypeIngress},
					},
				})
				i++
			}
		}
	}
	return policies
}

func (g *Generator) EgressPolicies(allowDNS bool) []*NetworkPolicy {
	var policies []*NetworkPolicy
	i := 0
	for _, egress := range g.EgressRuleSlices() {
		for _, target := range g.Targets {
			for _, ns := range g.Namespaces {
				if allowDNS {
					egress = append(egress, AllowDNSEgressRule)
				}
				policies = append(policies, &NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("policy-%d", i),
						Namespace: ns,
					},
					Spec: NetworkPolicySpec{
						PodSelector: target,
						Egress:      egress,
						PolicyTypes: []PolicyType{PolicyTypeEgress},
					},
				})
				i++
			}
		}
	}
	return policies
}

func (g *Generator) IngressEgressPolicies(allowDNS bool) []*NetworkPolicy {
	var policies []*NetworkPolicy
	i := 0
	for _, ingress := range g.IngressRuleSlices() {
		for _, egress := range g.EgressRuleSlices() {
			for _, target := range g.Targets {
				for _, ns := range g.Namespaces {
					if allowDNS {
						egress = append(egress, AllowDNSEgressRule)
					}
					policies = append(policies, &NetworkPolicy{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("policy-%d", i),
							Namespace: ns,
						},
						Spec: NetworkPolicySpec{
							PodSelector: target,
							Ingress:     ingress,
							Egress:      egress,
							PolicyTypes: []PolicyType{PolicyTypeIngress, PolicyTypeEgress},
						},
					})
					i++
				}
			}
		}
	}
	return policies
}

// multiple policies

func (g *Generator) IngressPolicySlices() [][]*NetworkPolicy {
	panic("TODO")
}

func (g *Generator) EgressPolicySlices() [][]*NetworkPolicy {
	panic("TODO")
}

func (g *Generator) IngressEgressPolicySlices() [][]*NetworkPolicy {
	panic("TODO")
}
