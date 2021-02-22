package generator

import (
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func basePolicy() *Netpol {
	return &Netpol{
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
}

type Setter func(policy *Netpol)

func SetPodSelector(sel metav1.LabelSelector) Setter {
	return func(policy *Netpol) {
		policy.Target.PodSelector = sel
	}
}

func SetRules(isIngress bool, rules []*Rule) Setter {
	return func(policy *Netpol) {
		if isIngress {
			policy.Ingress.Rules = rules
		} else {
			policy.Egress.Rules = rules
		}
	}
}

func SetPorts(isIngress bool, ports []NetworkPolicyPort) Setter {
	return func(policy *Netpol) {
		if isIngress {
			policy.Ingress.Rules[0].Ports = ports
		} else {
			policy.Egress.Rules[0].Ports = ports
		}
	}
}

func SetPeers(isIngress bool, peers []NetworkPolicyPeer) Setter {
	return func(policy *Netpol) {
		if isIngress {
			policy.Ingress.Rules[0].Peers = peers
		} else {
			policy.Egress.Rules[0].Peers = peers
		}
	}
}

func BuildPolicy(setters ...Setter) *Netpol {
	policy := basePolicy()
	for _, setter := range setters {
		setter(policy)
	}
	return policy
}

type BreadthGenerator struct {
	PodIP    string
	AllowDNS bool
}

func NewBreadthGenerator(allowDNS bool, podIP string) *BreadthGenerator {
	return &BreadthGenerator{
		PodIP:    podIP,
		AllowDNS: allowDNS,
	}
}

func (e *BreadthGenerator) Policies() []*Netpol {
	var policies []*Netpol

	// base policy
	policies = append(policies, BuildPolicy())

	// target
	// namespace
	//policies = append(policies, BuildPolicy(SetNamespace("y")))
	// pod selector
	for _, sel := range DefaultTargets() {
		policies = append(policies, BuildPolicy(SetPodSelector(sel)))
	}

	for _, isIngress := range []bool{true, false} {
		// empty rules
		policies = append(policies, BuildPolicy(SetRules(isIngress, []*Rule{})))

		// all ports/protocols
		policies = append(policies, BuildPolicy(SetPorts(isIngress, emptySliceOfPorts)))
		// specific protocol
		policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &tcp, Port: &port80}})))
		policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &udp, Port: &port80}})))
		policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &sctp, Port: &port80}})))
		// numbered port
		policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &tcp, Port: &port79}})))
		policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &sctp, Port: &port81}})))
		// named port
		policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &tcp, Port: &portServe79TCP}})))
		policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &udp, Port: &portServe81UDP}})))

		// TODO what's the expected outcome for these tests?
		// wrong protocol for port
		//policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &udp, Port: &portServe80TCP}})))
		//policies = append(policies, BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &sctp, Port: &portServe80TCP}})))

		// ns/pod peer, ipblock peer
		policies = append(policies, BuildPolicy(SetPeers(isIngress, emptySliceOfPeers)))
		for _, peers := range DefaultPeers(e.PodIP) {
			policies = append(policies, BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{peers})))
		}
	}

	// ingress/egress

	return policies
}

func (e *BreadthGenerator) GenerateTestCases() []*TestCase {
	var cases []*TestCase
	for _, policy := range e.Policies() {
		cases = append(cases, &TestCase{
			Description: "?",
			Steps:       []*TestStep{NewTestStep(ProbeAllAvailable, CreatePolicy(policy.NetworkPolicy()))},
		})
	}
	return cases
}
