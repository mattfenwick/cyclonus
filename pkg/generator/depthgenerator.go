package generator

import (
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var basePolicy = &Netpol{
	Name: "base",
	Target: &NetpolTarget{
		Namespace:   "x",
		PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{"pod": "a"}},
	},
	Ingress: &NetpolPeers{Rules: []*Rule{{
		Ports: []NetworkPolicyPort{{
			Port:     &port80,
			Protocol: &tcp,
		}},
		Peers: []NetworkPolicyPeer{{
			PodSelector:       podBCMatchExpressionsSelector,
			NamespaceSelector: nsXYMatchExpressionsSelector},
		}},
	}},
	Egress: &NetpolPeers{Rules: []*Rule{
		{
			Ports: []NetworkPolicyPort{{
				Port:     &port80,
				Protocol: &tcp,
			}},
			Peers: []NetworkPolicyPeer{{
				PodSelector:       podABMatchExpressionsSelector,
				NamespaceSelector: nsYZMatchExpressionsSelector},
			},
		},
		AllowDNSRule,
	},
	},
}

type DepthGenerator struct {
	PodIP    string
	AllowDNS bool
}

func NewDepthGenerator(allowDNS bool, podIP string) *DepthGenerator {
	return &DepthGenerator{
		PodIP:    podIP,
		AllowDNS: allowDNS,
	}
}

func (e *DepthGenerator) Policies() []*Netpol {
	var policies []*Netpol

	// TODO avoid duplicating breadth tests here?

	// base policy
	policies = append(policies, BuildPolicy())

	// target
	// namespace
	for _, ns := range DefaultNamespaces() {
		policies = append(policies, BuildPolicy(SetNamespace(ns)))
	}
	// pod selector
	for _, sel := range DefaultTargets() {
		policies = append(policies, BuildPolicy(SetPodSelector(sel)))
	}

	for _, isIngress := range []bool{true, false} {
		// empty rules
		policies = append(policies, BuildPolicy(SetRules(isIngress, []*Rule{})))

		// port/protocol
		policies = append(policies, BuildPolicy(SetPorts(isIngress, emptySliceOfPorts)))
		// different protocol
		policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &udp, Port: &port80}})))
		policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &sctp, Port: &port80}})))
		// different numbered port
		policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &tcp, Port: &port79}})))
		policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &tcp, Port: &port81}})))
		// different named port
		policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &tcp, Port: &portServe79TCP}})))
		policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &tcp, Port: &portServe80TCP}})))
		// wrong protocol for port
		policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &udp, Port: &portServe80TCP}})))
		policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &sctp, Port: &portServe80TCP}})))
		// pairs of ports
		for i, ports1 := range SinglePortProtocolTestCases() {
			policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{ports1})))
			for j, ports2 := range SinglePortProtocolTestCases() {
				if i < j {
					policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{ports1, ports2})))
				}
			}
		}

		// ns/pod peer, ipblock peer
		policies = append(policies, BuildPolicy(SetPeers(isIngress, emptySliceOfPeers)))
		for _, peers := range DefaultPeers(e.PodIP) {
			policies = append(policies, BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{peers})))
		}
	}

	// ingress/egress

	return policies
}

func (e *DepthGenerator) GenerateTestCases() []*TestCase {
	var cases []*TestCase
	for _, fp := range e.Policies() {
		cases = append(cases, NewSingleStepTestCase(fp.Description, ProbeAllAvailable, CreatePolicy(fp.NetworkPolicy())))
	}
	return cases
}
