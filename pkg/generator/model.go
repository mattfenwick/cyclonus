package generator

import (
	"github.com/pkg/errors"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Netpol helps us to avoid the To/From Ingress/Egress dance.  By splitting a NetworkPolicy into
// Target and Peers, it makes them easier to manipulate.
type Netpol struct {
	Name    string
	Target  *NetpolTarget
	Ingress *NetpolPeers
	Egress  *NetpolPeers
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
