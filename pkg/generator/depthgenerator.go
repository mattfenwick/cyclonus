package generator

import (
	. "k8s.io/api/networking/v1"
)

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
		cases = append(cases, NewSingleStepTestCase(fp.Description, NewStringSet(), ProbeAllAvailable, CreatePolicy(fp.NetworkPolicy())))
	}
	return cases
}
