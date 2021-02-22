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

type Piece func(policy *Netpol) map[string]bool

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

func SetPodSelector(sel metav1.LabelSelector) Piece {
	return func(policy *Netpol) map[string]bool {
		policy.Target.PodSelector = sel
		return targetPodSelectorFeatures(sel)
	}
}

func SetRules(isIngress bool, rules []*Rule) Piece {
	return func(policy *Netpol) map[string]bool {
		if isIngress {
			policy.Ingress.Rules = rules
			var kubeRules []NetworkPolicyIngressRule
			for _, rule := range rules {
				kubeRules = append(kubeRules, rule.Ingress())
			}
			return addDirectionalityToKeys(isIngress, ingressFeatures(kubeRules))
		} else {
			policy.Egress.Rules = rules
			var kubeRules []NetworkPolicyEgressRule
			for _, rule := range rules {
				kubeRules = append(kubeRules, rule.Egress())
			}
			return addDirectionalityToKeys(isIngress, egressFeatures(kubeRules))
		}
	}
}

func SetPorts(isIngress bool, ports []NetworkPolicyPort) Piece {
	return func(policy *Netpol) map[string]bool {
		if isIngress {
			policy.Ingress.Rules[0].Ports = ports
		} else {
			policy.Egress.Rules[0].Ports = ports
		}
		return addDirectionalityToKeys(isIngress, portFeatures(ports))
	}
}

func SetPeers(isIngress bool, peers []NetworkPolicyPeer) Piece {
	return func(policy *Netpol) map[string]bool {
		if isIngress {
			policy.Ingress.Rules[0].Peers = peers
		} else {
			policy.Egress.Rules[0].Peers = peers
		}
		return addDirectionalityToKeys(isIngress, peerFeatures(peers))
	}
}

func BuildPolicy(pieces ...Piece) *FeaturePolicy {
	policy := copyPolicy(basePolicy)
	features := map[string]bool{}
	for _, p := range pieces {
		features = mergeSets(features, p(policy))
	}
	return &FeaturePolicy{Features: features, Policy: policy}
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

type FeaturePolicy struct {
	Features map[string]bool
	Policy   *Netpol
}

func (e *BreadthGenerator) Policies() []*FeaturePolicy {
	var policies []*FeaturePolicy

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
	for _, fp := range e.Policies() {
		cases = append(cases, &TestCase{
			Description: "?",
			Features:    fp.Features,
			Steps:       []*TestStep{NewTestStep(ProbeAllAvailable, CreatePolicy(fp.Policy.NetworkPolicy()))},
		})
	}
	return cases
}
