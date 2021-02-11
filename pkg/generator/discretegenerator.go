package generator

import (
	"fmt"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewDefaultDiscreteGenerator(allowDNS bool, podIP string) *DiscreteGenerator {
	return &DiscreteGenerator{
		AllowDNS:   allowDNS,
		Ports:      []NetworkPolicyPort{emptyPort, sctpOnAnyPort, implicitTCPOnPort80, explicitUDPOnPort80, namedPort81TPCP},
		PodPeers:   DefaultPeers(podIP),
		Targets:    DefaultTargets(),
		Namespaces: DefaultNamespaces(),
		// ingress
		TypicalIngressPorts:     []NetworkPolicyPort{implicitTCPOnPort80},
		TypicalIngressPeers:     []NetworkPolicyPeer{{PodSelector: podBCMatchExpressionsSelector, NamespaceSelector: nsYZMatchExpressionsSelector}},
		TypicalIngressTarget:    []metav1.LabelSelector{*podABMatchExpressionsSelector},
		TypicalIngressNamespace: []string{"x"},
		// egress
		TypicalEgressPorts:     []NetworkPolicyPort{namedPort81TPCP},
		TypicalEgressPeers:     []NetworkPolicyPeer{{PodSelector: podABMatchExpressionsSelector, NamespaceSelector: nsXYMatchExpressionsSelector}},
		TypicalEgressTarget:    []metav1.LabelSelector{*podCMatchLabelsSelector},
		TypicalEgressNamespace: []string{"z"},
	}
}

// we want a background "typical" policy that will always pass on a CNI with an expected connectivity table,
//   but also will allow some traffic but deny others.  Then the perturbations to that base policy should
//   affect other pods.
type DiscreteGenerator struct {
	AllowDNS bool

	Ports      []NetworkPolicyPort
	PodPeers   []NetworkPolicyPeer
	Targets    []metav1.LabelSelector
	Namespaces []string

	TypicalIngressPorts     []NetworkPolicyPort
	TypicalIngressPeers     []NetworkPolicyPeer
	TypicalIngressTarget    []metav1.LabelSelector
	TypicalIngressNamespace []string

	TypicalEgressPorts     []NetworkPolicyPort
	TypicalEgressPeers     []NetworkPolicyPeer
	TypicalEgressTarget    []metav1.LabelSelector
	TypicalEgressNamespace []string
}

func (g *DiscreteGenerator) PortSlices() [][]NetworkPolicyPort {
	// length 0
	slices := [][]NetworkPolicyPort{nil, emptySliceOfPorts}
	// length 1
	for _, s := range g.Ports {
		slices = append(slices, []NetworkPolicyPort{s})
	}
	return slices
}

func (g *DiscreteGenerator) PeerSlices() [][]NetworkPolicyPeer {
	// 0 length
	slices := [][]NetworkPolicyPeer{nil, emptySliceOfPeers}
	// 1 length
	for _, p := range g.PodPeers {
		slices = append(slices, []NetworkPolicyPeer{p})
	}
	return slices
}

func (g *DiscreteGenerator) Rules() []*Rule {
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

func (g *DiscreteGenerator) RuleSlices() [][]*Rule {
	// 0 length
	slices := [][]*Rule{nil, emptySliceOfRules}
	// 1 length
	for _, ingress := range g.Rules() {
		slices = append(slices, []*Rule{ingress})
	}
	return slices
}

// single policies, unidimensional generation

func targets(nss []string, selectors []metav1.LabelSelector) []*NetpolTarget {
	var ts []*NetpolTarget
	for _, ns := range nss {
		for _, sel := range selectors {
			ts = append(ts, &NetpolTarget{Namespace: ns, PodSelector: sel})
		}
	}
	return ts
}

func netpolPeers(ports []NetworkPolicyPort, peers []NetworkPolicyPeer) []*NetpolPeers {
	var ps []*NetpolPeers
	for _, port := range ports {
		for _, peer := range peers {
			ps = append(ps, &NetpolPeers{Rules: []*Rule{{Ports: []NetworkPolicyPort{port}, Peers: []NetworkPolicyPeer{peer}}}})
		}
	}
	return ps
}

func ingress(nss []string, selectors []metav1.LabelSelector, ports []NetworkPolicyPort, peers []NetworkPolicyPeer) []*Netpol {
	var policies []*Netpol
	for _, target := range targets(nss, selectors) {
		for _, peer := range netpolPeers(ports, peers) {
			policies = append(policies, &Netpol{
				Target:  target,
				Ingress: peer,
			})
		}
	}
	return policies
}

func egress(allowDNS bool, nss []string, selectors []metav1.LabelSelector, ports []NetworkPolicyPort, peers []NetworkPolicyPeer) []*Netpol {
	var policies []*Netpol
	for _, target := range targets(nss, selectors) {
		for _, peer := range netpolPeers(ports, peers) {
			if allowDNS {
				peer.Rules = append(peer.Rules, AllowDNSRule)
			}
			policies = append(policies, &Netpol{
				Target: target,
				Egress: peer,
			})
		}
	}
	return policies
}

func ingressEgress(
	allowDNS bool,
	count *int,
	iNss []string,
	iSel []metav1.LabelSelector,
	iPorts []NetworkPolicyPort,
	iPeers []NetworkPolicyPeer,
	eNss []string,
	eSel []metav1.LabelSelector,
	ePorts []NetworkPolicyPort,
	ePeers []NetworkPolicyPeer) [][]*Netpol {
	var policies [][]*Netpol
	for i, ing := range ingress(iNss, iSel, iPorts, iPeers) {
		for j, eg := range egress(allowDNS, eNss, eSel, ePorts, ePeers) {
			ing.Name = fmt.Sprintf("ingress-%d-%d-%d", *count, i, j)
			eg.Name = fmt.Sprintf("egress-%d-%d-%d", *count, i, j)
			policies = append(policies, []*Netpol{ing, eg})
			*count++
		}
	}
	return policies
}

func (g *DiscreteGenerator) fragmentPolicies() [][]*Netpol {
	var policies [][]*Netpol
	count := 0
	policies = append(policies, ingressEgress(g.AllowDNS, &count, g.TypicalIngressNamespace, g.TypicalIngressTarget, g.TypicalIngressPorts, g.TypicalIngressPeers, g.TypicalEgressNamespace, g.TypicalEgressTarget, g.TypicalEgressPorts, g.TypicalEgressPeers)...)

	policies = append(policies, ingressEgress(g.AllowDNS, &count, g.Namespaces, g.TypicalIngressTarget, g.TypicalIngressPorts, g.TypicalIngressPeers, g.TypicalEgressNamespace, g.TypicalEgressTarget, g.TypicalEgressPorts, g.TypicalEgressPeers)...)
	policies = append(policies, ingressEgress(g.AllowDNS, &count, g.TypicalIngressNamespace, g.Targets, g.TypicalIngressPorts, g.TypicalIngressPeers, g.TypicalEgressNamespace, g.TypicalEgressTarget, g.TypicalEgressPorts, g.TypicalEgressPeers)...)
	policies = append(policies, ingressEgress(g.AllowDNS, &count, g.TypicalIngressNamespace, g.TypicalIngressTarget, g.Ports, g.TypicalIngressPeers, g.TypicalEgressNamespace, g.TypicalEgressTarget, g.TypicalEgressPorts, g.TypicalEgressPeers)...)
	policies = append(policies, ingressEgress(g.AllowDNS, &count, g.TypicalIngressNamespace, g.TypicalIngressTarget, g.TypicalIngressPorts, g.PodPeers, g.TypicalEgressNamespace, g.TypicalEgressTarget, g.TypicalEgressPorts, g.TypicalEgressPeers)...)

	policies = append(policies, ingressEgress(g.AllowDNS, &count, g.TypicalIngressNamespace, g.TypicalIngressTarget, g.TypicalIngressPorts, g.TypicalIngressPeers, g.Namespaces, g.TypicalEgressTarget, g.TypicalEgressPorts, g.TypicalEgressPeers)...)
	policies = append(policies, ingressEgress(g.AllowDNS, &count, g.TypicalIngressNamespace, g.TypicalIngressTarget, g.TypicalIngressPorts, g.TypicalIngressPeers, g.TypicalEgressNamespace, g.Targets, g.TypicalEgressPorts, g.TypicalEgressPeers)...)
	policies = append(policies, ingressEgress(g.AllowDNS, &count, g.TypicalIngressNamespace, g.TypicalIngressTarget, g.TypicalIngressPorts, g.TypicalIngressPeers, g.TypicalEgressNamespace, g.TypicalEgressTarget, g.Ports, g.TypicalEgressPeers)...)
	policies = append(policies, ingressEgress(g.AllowDNS, &count, g.TypicalIngressNamespace, g.TypicalIngressTarget, g.TypicalIngressPorts, g.TypicalIngressPeers, g.TypicalEgressNamespace, g.TypicalEgressTarget, g.TypicalEgressPorts, g.PodPeers)...)

	return policies
}

func (g *DiscreteGenerator) GenerateTestCases() []*TestCase {
	var testCases []*TestCase
	for _, netpols := range g.fragmentPolicies() {
		var actions []*Action
		for _, np := range netpols {
			actions = append(actions, CreatePolicy(np.NetworkPolicy()))
		}
		testCases = append(testCases, NewSingleStepTestCase("TODO", allAvailable, actions...))
	}
	return testCases
}
