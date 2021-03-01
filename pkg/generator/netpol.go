package generator

import (
	"github.com/pkg/errors"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Netpol helps us to avoid the To/From Ingress/Egress dance.  By splitting a NetworkPolicy into
// Target and Peers, it makes them easier to manipulate.
type Netpol struct {
	Name        string
	Description string
	Target      *NetpolTarget
	Ingress     *NetpolPeers
	Egress      *NetpolPeers
}

func NewNetpol(policy *NetworkPolicy) *Netpol {
	var ingress, egress = &NetpolPeers{}, &NetpolPeers{}
	for _, i := range policy.Spec.Ingress {
		ingress.Rules = append(ingress.Rules, &Rule{
			Ports: i.Ports,
			Peers: i.From,
		})
	}
	for _, i := range policy.Spec.Egress {
		egress.Rules = append(egress.Rules, &Rule{
			Ports: i.Ports,
			Peers: i.To,
		})
	}
	return &Netpol{
		Name:        policy.Namespace,
		Description: "generated from networkingv1.NetworkPolicy",
		Target: &NetpolTarget{
			Namespace:   policy.Namespace,
			PodSelector: policy.Spec.PodSelector,
		},
		Ingress: ingress,
		Egress:  egress,
	}
}

func (n *Netpol) NetworkPolicy() *NetworkPolicy {
	return &NetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NetworkPolicy",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      n.Name,
			Namespace: n.Target.Namespace,
		},
		Spec: *n.NetworkPolicySpec(),
	}
}

func (n *Netpol) NetworkPolicySpec() *NetworkPolicySpec {
	var types []PolicyType
	var ingress []NetworkPolicyIngressRule
	var egress []NetworkPolicyEgressRule
	if n.Ingress != nil {
		types = append(types, PolicyTypeIngress)
		for _, rule := range n.Ingress.Rules {
			ingress = append(ingress, rule.Ingress())
		}
	}
	if n.Egress != nil {
		types = append(types, PolicyTypeEgress)
		for _, rule := range n.Egress.Rules {
			egress = append(egress, rule.Egress())
		}
	}
	if len(types) == 0 {
		panic(errors.Errorf("cannot have 0 policy types"))
	}
	return &NetworkPolicySpec{
		PodSelector: n.Target.PodSelector,
		Ingress:     ingress,
		Egress:      egress,
		PolicyTypes: types,
	}
}

type NetpolTarget struct {
	Namespace   string
	PodSelector metav1.LabelSelector
}

func NewNetpolTarget(namespace string, matchLabels map[string]string, matchExpressions []metav1.LabelSelectorRequirement) *NetpolTarget {
	return &NetpolTarget{
		Namespace: namespace,
		PodSelector: metav1.LabelSelector{
			MatchLabels:      matchLabels,
			MatchExpressions: matchExpressions,
		},
	}
}

type NetpolPeers struct {
	Rules []*Rule
}

type Rule struct {
	Ports []NetworkPolicyPort
	Peers []NetworkPolicyPeer
}

//func RulesFromPortsAndPeers(ports []NetworkPolicyPort, peers []NetworkPolicyPeer) []*Rule {
//	var rules []*Rule
//	for _, port := range ports {
//		for _, peer := range peers {
//			rules = append(rules, &Rule{
//				Ports: port,
//				Peers: peer,
//			})
//		}
//	}
//}

func (r *Rule) Ingress() NetworkPolicyIngressRule {
	return NetworkPolicyIngressRule{
		Ports: r.Ports,
		From:  r.Peers,
	}
}

func (r *Rule) Egress() NetworkPolicyEgressRule {
	return NetworkPolicyEgressRule{
		Ports: r.Ports,
		To:    r.Peers,
	}
}

// Setter is used to declaratively build network policies
type Setter func(policy *Netpol)

func SetDescription(description string) Setter {
	return func(policy *Netpol) {
		policy.Description = description
	}
}

func SetNamespace(ns string) Setter {
	return func(policy *Netpol) {
		policy.Target.Namespace = ns
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

func SetPodSelector(sel metav1.LabelSelector) Setter {
	return func(policy *Netpol) {
		policy.Target.PodSelector = sel
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
	policy := baseTestPolicy()
	for _, setter := range setters {
		setter(policy)
	}
	return policy
}

func baseTestPolicy() *Netpol {
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
		}},
	}
}
