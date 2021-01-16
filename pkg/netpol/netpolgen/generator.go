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

func (g *Generator) Rules() []*Rule {
	var rules []*Rule
	for _, ports := range g.PortSlices() {
		for _, peers := range g.PeerSlices() {
			rules = append(rules, &Rule{
				Ports: ports,
				Peers: peers,
			})
		}
	}
	return rules
}

func (g *Generator) RuleSlices() [][]*Rule {
	// 0 length
	slices := [][]*Rule{nil, emptySliceOfRules}
	// 1 length
	for _, ingress := range g.Rules() {
		slices = append(slices, []*Rule{ingress})
	}
	// 2 length
	for i, ing1 := range g.Rules() {
		for j, ing2 := range g.Rules() {
			if i < j {
				slices = append(slices, []*Rule{ing1, ing2})
			}
		}
	}
	return slices
}

// single policies, unidimensional generation

func (g *Generator) varyPolicies(count *int, isIngress bool, nss []string, targets []metav1.LabelSelector, ports [][]NetworkPolicyPort, peers [][]NetworkPolicyPeer) []*NetworkPolicy {
	var policies []*NetworkPolicy
	for i, ns := range nss {
		for j, target := range targets {
			for k, port := range ports {
				for l, peer := range peers {
					if isIngress {
						name := fmt.Sprintf("vary-ingress-%d-%d-%d-%d-%d", *count, i, j, k, l)
						policies = append(policies, (&Netpol{Name: name, Namespace: ns, PodSelector: target, IngressRules: []*Rule{{Ports: port, Peers: peer}}}).NetworkPolicy())
					} else {
						name := fmt.Sprintf("vary-egress-%d-%d-%d-%d-%d", *count, i, j, k, l)
						policies = append(policies, (&Netpol{Name: name, Namespace: ns, PodSelector: target, EgressRules: []*Rule{{Ports: port, Peers: peer}}}).NetworkPolicy())
					}
					*count++
				}
			}
		}
	}
	return policies
}

func (g *Generator) VaryIngressPolicies() []*NetworkPolicy {
	isIngress := true
	var policies []*NetworkPolicy
	count := 0
	policies = append(policies, g.varyPolicies(&count, isIngress, g.Namespaces, []metav1.LabelSelector{g.TypicalTarget}, [][]NetworkPolicyPort{g.TypicalPorts}, [][]NetworkPolicyPeer{g.TypicalPeers})...)
	policies = append(policies, g.varyPolicies(&count, isIngress, []string{g.TypicalNamespace}, g.Targets, [][]NetworkPolicyPort{g.TypicalPorts}, [][]NetworkPolicyPeer{g.TypicalPeers})...)
	policies = append(policies, g.varyPolicies(&count, isIngress, []string{g.TypicalNamespace}, []metav1.LabelSelector{g.TypicalTarget}, g.PortSlices(), [][]NetworkPolicyPeer{g.TypicalPeers})...)
	policies = append(policies, g.varyPolicies(&count, isIngress, []string{g.TypicalNamespace}, []metav1.LabelSelector{g.TypicalTarget}, [][]NetworkPolicyPort{g.TypicalPorts}, g.PeerSlices())...)
	return policies
}

func (g *Generator) VaryEgressPolicies(allowDNS bool) []*NetworkPolicy {
	isIngress := false
	var policies []*NetworkPolicy
	count := 0
	policies = append(policies, g.varyPolicies(&count, isIngress, g.Namespaces, []metav1.LabelSelector{g.TypicalTarget}, [][]NetworkPolicyPort{g.TypicalPorts}, [][]NetworkPolicyPeer{g.TypicalPeers})...)
	policies = append(policies, g.varyPolicies(&count, isIngress, []string{g.TypicalNamespace}, g.Targets, [][]NetworkPolicyPort{g.TypicalPorts}, [][]NetworkPolicyPeer{g.TypicalPeers})...)
	policies = append(policies, g.varyPolicies(&count, isIngress, []string{g.TypicalNamespace}, []metav1.LabelSelector{g.TypicalTarget}, g.PortSlices(), [][]NetworkPolicyPeer{g.TypicalPeers})...)
	policies = append(policies, g.varyPolicies(&count, isIngress, []string{g.TypicalNamespace}, []metav1.LabelSelector{g.TypicalTarget}, [][]NetworkPolicyPort{g.TypicalPorts}, g.PeerSlices())...)
	if allowDNS {
		for _, pol := range policies {
			pol.Spec.Egress = append(pol.Spec.Egress, AllowDNSEgressRule)
		}
	}
	return policies
}

// single policies, multidimensional generation

func (g *Generator) multidimensionalPolicies(isIngress bool, allowDNS bool) []*NetworkPolicy {
	var policies []*NetworkPolicy
	i := 0
	for _, rules := range g.RuleSlices() {
		for _, target := range g.Targets {
			for _, ns := range g.Namespaces {
				var ingresses []*Rule
				var egresses []*Rule
				for _, rule := range rules {
					if isIngress {
						ingresses = append(ingresses, rule)
					} else {
						egresses = append(egresses, rule)
					}
				}
				if !isIngress && allowDNS {
					egresses = append(egresses, AllowDNSRule)
				}
				policies = append(policies, (&Netpol{
					Name:         fmt.Sprintf("policy-%d", i),
					Namespace:    ns,
					PodSelector:  target,
					IngressRules: ingresses,
					EgressRules:  egresses,
				}).NetworkPolicy())
				i++
			}
		}
	}
	return policies
}

func (g *Generator) IngressPolicies() []*NetworkPolicy {
	return g.multidimensionalPolicies(true, false)
}

func (g *Generator) EgressPolicies(allowDNS bool) []*NetworkPolicy {
	return g.multidimensionalPolicies(false, allowDNS)
}

func (g *Generator) IngressEgressPolicies(allowDNS bool) []*NetworkPolicy {
	panic("TODO -- how to get this to a workable number?")
	//var policies []*NetworkPolicy
	//i := 0
	//for _, ingress := range g.IngressRuleSlices() {
	//	for _, egress := range g.EgressRuleSlices() {
	//		for _, target := range g.Targets {
	//			for _, ns := range g.Namespaces {
	//				if allowDNS {
	//					egress = append(egress, AllowDNSEgressRule)
	//				}
	//				policies = append(policies, &NetworkPolicy{
	//					ObjectMeta: metav1.ObjectMeta{
	//						Name:      fmt.Sprintf("policy-%d", i),
	//						Namespace: ns,
	//					},
	//					Spec: NetworkPolicySpec{
	//						PodSelector: target,
	//						Ingress:     ingress,
	//						Egress:      egress,
	//						PolicyTypes: []PolicyType{PolicyTypeIngress, PolicyTypeEgress},
	//					},
	//				})
	//				i++
	//			}
	//		}
	//	}
	//}
	//return policies
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
