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
	Egress: nil,
}

type Piece func(policy *Netpol) *Features

func copyPolicy(policy *Netpol) *Netpol {
	return &Netpol{
		Name: policy.Name,
		Target: &NetpolTarget{
			Namespace:   policy.Target.Namespace,
			PodSelector: policy.Target.PodSelector,
		},
		Ingress: copyPeers(policy.Ingress),
		Egress:  copyPeers(policy.Egress),
	}
}

func copyPeers(peers *NetpolPeers) *NetpolPeers {
	if peers == nil {
		return nil
	}
	newPeers := &NetpolPeers{Rules: []*Rule{}}
	for _, rule := range peers.Rules {
		newRule := &Rule{}
		for _, peer := range rule.Peers {
			newRule.Peers = append(newRule.Peers, peer)
		}
		for _, port := range rule.Ports {
			newRule.Ports = append(newRule.Ports, port)
		}
		newPeers.Rules = append(newPeers.Rules, newRule)
	}
	return newPeers
}

func SetNamespace(ns string) Piece {
	return func(policy *Netpol) *Features {
		policy.Target.Namespace = ns
		if ns == "" {
			return &Features{Target: map[TargetFeature]bool{TargetFeatureNamespaceEmpty: true}}
		}
		return &Features{}
	}
}

//func SetIngress() Piece {
//	return func(policy *Netpol) {
//		policy.Ingress.Rules = append(policy.Ingress.Rules, &Rule{
//			Ports: []NetworkPolicyPort{{}},
//			Peers: []NetworkPolicyPeer{{}},
//		})
//	}
//}

func SetPodSelector(sel metav1.LabelSelector) Piece {
	return func(policy *Netpol) *Features {
		policy.Target.PodSelector = sel
		return &Features{Target: targetPodSelectorFeatures(sel)}
	}
}

//func SetName(name string) Piece {
//	return func(policy *Netpol) {
//		policy.Name = name
//	}
//}

//func SetIngressRules(rules []*Rule) Piece {
//	return func(policy *Netpol) *Features {
//		policy.Ingress.Rules = rules
//		return &Features{Ingress: ruleFeatures(ports)}
//	}
//}

func SetIngressPorts(ports []NetworkPolicyPort) Piece {
	return func(policy *Netpol) *Features {
		policy.Ingress.Rules[0].Ports = ports
		return &Features{Ingress: portFeatures(ports)}
	}
}

func SetIngressPeers(peers []NetworkPolicyPeer) Piece {
	return func(policy *Netpol) *Features {
		policy.Ingress.Rules[0].Peers = peers
		return &Features{Ingress: peerFeatures(peers)}
	}
}

//func SetProtocol(protocol v1.Protocol) Piece {
//	return func(policy *Netpol) {
//		policy.Ingress.Rules[0].Ports[0].Protocol = &protocol
//	}
//}
//
//func SetPeerPodSelector(sel *metav1.LabelSelector) Piece {
//	return func(policy *Netpol) {
//		policy.Ingress.Rules[0].Peers[0].PodSelector = sel
//	}
//}
//
//func SetPeerNamespaceSelector(sel *metav1.LabelSelector) Piece {
//	return func(policy *Netpol) {
//		policy.Ingress.Rules[0].Peers[0].NamespaceSelector = sel
//	}
//}
//
//func SetPeerIPBlock(ipBlock *IPBlock) Piece {
//	return func(policy *Netpol) {
//		policy.Ingress.Rules[0].Peers[0].IPBlock = ipBlock
//	}
//}

func BuildPolicy(pieces ...Piece) *FeaturePolicy {
	policy := copyPolicy(basePolicy)
	features := &Features{}
	for _, p := range pieces {
		features = features.Combine(p(policy))
	}
	return &FeaturePolicy{Features: features, Policy: policy}
}

type ExperimentalGenerator struct {
	PodIP    string
	AllowDNS bool
}

func NewExperimentalGenerator(allowDNS bool, podIP string) *ExperimentalGenerator {
	return &ExperimentalGenerator{
		PodIP:    podIP,
		AllowDNS: allowDNS,
	}
}

type FeaturePolicy struct {
	Features *Features
	Policy   *Netpol
}

func experimentalPorts() []NetworkPolicyPort {
	return []NetworkPolicyPort{
		{Protocol: nil, Port: nil},
		{Protocol: &tcp, Port: nil},
		{Protocol: nil, Port: &port80},
		{Protocol: &tcp, Port: &port80},
		{Protocol: nil, Port: &portServe81TCP},
		{Protocol: &tcp, Port: &portServe81TCP},
	}
}

func (e *ExperimentalGenerator) Policies() []*FeaturePolicy {
	var policies []*FeaturePolicy

	// base policy
	policies = append(policies, BuildPolicy())

	// namespace
	for _, ns := range DefaultNamespaces() {
		policies = append(policies, BuildPolicy(SetNamespace(ns)))
	}

	// pod selector
	for _, sel := range DefaultTargets() {
		policies = append(policies, BuildPolicy(SetPodSelector(sel)))
	}

	// port
	policies = append(policies, BuildPolicy(SetIngressPorts(emptySliceOfPorts)))
	// different protocol
	policies = append(policies, BuildPolicy(SetIngressPorts([]NetworkPolicyPort{{Protocol: &udp, Port: &port80}})))
	policies = append(policies, BuildPolicy(SetIngressPorts([]NetworkPolicyPort{{Protocol: &sctp, Port: &port80}})))
	// different numbered port
	policies = append(policies, BuildPolicy(SetIngressPorts([]NetworkPolicyPort{{Protocol: &tcp, Port: &port79}})))
	policies = append(policies, BuildPolicy(SetIngressPorts([]NetworkPolicyPort{{Protocol: &tcp, Port: &port81}})))
	// different named port
	policies = append(policies, BuildPolicy(SetIngressPorts([]NetworkPolicyPort{{Protocol: &tcp, Port: &portServe79TCP}})))
	policies = append(policies, BuildPolicy(SetIngressPorts([]NetworkPolicyPort{{Protocol: &tcp, Port: &portServe80TCP}})))
	// wrong protocol for port
	policies = append(policies, BuildPolicy(SetIngressPorts([]NetworkPolicyPort{{Protocol: &udp, Port: &portServe80TCP}})))
	policies = append(policies, BuildPolicy(SetIngressPorts([]NetworkPolicyPort{{Protocol: &sctp, Port: &portServe80TCP}})))
	// pairs of ports
	for i, ports1 := range experimentalPorts() {
		policies = append(policies, BuildPolicy(SetIngressPorts([]NetworkPolicyPort{ports1})))
		for j, ports2 := range experimentalPorts() {
			if i < j {
				policies = append(policies, BuildPolicy(SetIngressPorts([]NetworkPolicyPort{ports1, ports2})))
			}
		}
	}

	// ns/pod peer, ipblock peer
	policies = append(policies, BuildPolicy(SetIngressPeers(emptySliceOfPeers)))
	for _, peers := range DefaultPeers(e.PodIP) {
		policies = append(policies, BuildPolicy(SetIngressPeers([]NetworkPolicyPeer{peers})))
	}

	// ingress/egress

	return policies
}

func (e *ExperimentalGenerator) GenerateTestCases() []*TestCase {
	var cases []*TestCase
	for _, fp := range e.Policies() {
		//cases = append(cases, NewSingleStepTestCase("?", allAvailable, CreatePolicy(fp.Policy.NetworkPolicy())))
		cases = append(cases, &TestCase{
			Description: "?",
			Features:    fp.Features,
			Steps:       []*TestStep{NewTestStep(allAvailable, CreatePolicy(fp.Policy.NetworkPolicy()))},
		})
	}
	return cases
}
