package generator

import (
	"fmt"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewDefaultFragmentGenerator(namespaces []string, podIP string) *FragmentGenerator {
	return &FragmentGenerator{
		Ports:            DefaultPorts(),
		PodPeers:         DefaultPodPeers(podIP),
		Targets:          DefaultTargets(),
		Namespaces:       namespaces,
		TypicalPorts:     TypicalPorts,
		TypicalPeers:     TypicalPeers,
		TypicalTarget:    TypicalTarget,
		TypicalNamespace: TypicalNamespace,
	}
}

type FragmentGenerator struct {
	// multidimensional generation
	Ports      []NetworkPolicyPort
	PodPeers   []NetworkPolicyPeer
	Targets    []metav1.LabelSelector
	Namespaces []string
	// unidimensional typicals
	TypicalPorts     []NetworkPolicyPort
	TypicalPeers     []NetworkPolicyPeer
	TypicalTarget    metav1.LabelSelector
	TypicalNamespace string
}

func (g *FragmentGenerator) PortSlices() [][]NetworkPolicyPort {
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

func (g *FragmentGenerator) PeerSlices() [][]NetworkPolicyPeer {
	// 0 length
	slices := [][]NetworkPolicyPeer{nil, emptySliceOfPeers}
	// 1 length
	for _, p := range g.PodPeers {
		slices = append(slices, []NetworkPolicyPeer{p})
	}
	// 2 length
	for i, p1 := range g.PodPeers {
		for j, p2 := range g.PodPeers {
			if i < j {
				slices = append(slices, []NetworkPolicyPeer{p1, p2})
			}
		}
	}
	return slices
}

func (g *FragmentGenerator) Rules() []*Rule {
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

func (g *FragmentGenerator) RuleSlices() [][]*Rule {
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

func (g *FragmentGenerator) fragmentPolicies(count *int, isIngress bool, nss []string, targets []metav1.LabelSelector, ports [][]NetworkPolicyPort, peers [][]NetworkPolicyPeer) []*NetworkPolicy {
	var policies []*NetworkPolicy
	for i, ns := range nss {
		for j, target := range targets {
			for k, port := range ports {
				for l, peer := range peers {
					var ingress, egress *NetpolPeers
					var desc string
					if isIngress {
						desc = "ingress"
						ingress = &NetpolPeers{Rules: []*Rule{{Ports: port, Peers: peer}}}
					} else {
						desc = "egress"
						egress = &NetpolPeers{Rules: []*Rule{{Ports: port, Peers: peer}}}
					}
					name := fmt.Sprintf("fragment-%s-%d-%d-%d-%d-%d", desc, *count, i, j, k, l)
					policies = append(policies, (&Netpol{
						Name:    name,
						Target:  &NetpolTarget{Namespace: ns, PodSelector: target},
						Ingress: ingress,
						Egress:  egress}).NetworkPolicy())
					*count++
				}
			}
		}
	}
	return policies
}

func (g *FragmentGenerator) fragmentPoliciesWrapper(isIngress bool) []*NetworkPolicy {
	var policies []*NetworkPolicy
	count := 0
	policies = append(policies, g.fragmentPolicies(&count, isIngress, g.Namespaces, []metav1.LabelSelector{g.TypicalTarget}, [][]NetworkPolicyPort{g.TypicalPorts}, [][]NetworkPolicyPeer{g.TypicalPeers})...)
	policies = append(policies, g.fragmentPolicies(&count, isIngress, []string{g.TypicalNamespace}, g.Targets, [][]NetworkPolicyPort{g.TypicalPorts}, [][]NetworkPolicyPeer{g.TypicalPeers})...)
	policies = append(policies, g.fragmentPolicies(&count, isIngress, []string{g.TypicalNamespace}, []metav1.LabelSelector{g.TypicalTarget}, g.PortSlices(), [][]NetworkPolicyPeer{g.TypicalPeers})...)
	policies = append(policies, g.fragmentPolicies(&count, isIngress, []string{g.TypicalNamespace}, []metav1.LabelSelector{g.TypicalTarget}, [][]NetworkPolicyPort{g.TypicalPorts}, g.PeerSlices())...)
	return policies
}

func (g *FragmentGenerator) FragmentIngressPolicies() []*NetworkPolicy {
	policies := g.fragmentPoliciesWrapper(true)
	// special case: empty ingress/egress
	return append(policies, (&Netpol{
		Name:    "fragment-ingress-empty",
		Target:  &NetpolTarget{Namespace: g.TypicalNamespace, PodSelector: g.TypicalTarget},
		Ingress: &NetpolPeers{}}).NetworkPolicy())
}

func (g *FragmentGenerator) FragmentEgressPolicies(allowDNS bool) []*NetworkPolicy {
	policies := g.fragmentPoliciesWrapper(false)
	if allowDNS {
		for _, pol := range policies {
			pol.Spec.Egress = append(pol.Spec.Egress, AllowDNSRule.Egress())
		}
	}
	// special case: empty ingress/egress
	return append(policies, (&Netpol{
		Name:   "fragment-egress-empty",
		Target: &NetpolTarget{Namespace: g.TypicalNamespace, PodSelector: g.TypicalTarget},
		Egress: &NetpolPeers{}}).NetworkPolicy())
}

func (g *FragmentGenerator) FragmentPolicies(allowDNS bool) []*NetworkPolicy {
	return append(g.FragmentIngressPolicies(), g.FragmentEgressPolicies(allowDNS)...)
}

// single policies, multidimensional generation

func (g *FragmentGenerator) multidimensionalPolicies(isIngress bool, allowDNS bool) []*NetworkPolicy {
	var policies []*NetworkPolicy
	i := 0
	for _, rules := range g.RuleSlices() {
		for _, target := range g.Targets {
			for _, ns := range g.Namespaces {
				var ingress, egress *NetpolPeers
				if isIngress {
					ingress = &NetpolPeers{Rules: rules}
				} else {
					egress = &NetpolPeers{Rules: rules}
				}
				if !isIngress && allowDNS {
					egress.Rules = append(egress.Rules, AllowDNSRule)
				}
				policies = append(policies, (&Netpol{
					Name:    fmt.Sprintf("policy-%d", i),
					Target:  &NetpolTarget{Namespace: ns, PodSelector: target},
					Ingress: ingress,
					Egress:  egress,
				}).NetworkPolicy())
				i++
			}
		}
	}
	return policies
}

func (g *FragmentGenerator) IngressPolicies() []*NetworkPolicy {
	return g.multidimensionalPolicies(true, false)
}

func (g *FragmentGenerator) EgressPolicies(allowDNS bool) []*NetworkPolicy {
	return g.multidimensionalPolicies(false, allowDNS)
}

func (g *FragmentGenerator) IngressEgressPolicies(allowDNS bool) []*NetworkPolicy {
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
