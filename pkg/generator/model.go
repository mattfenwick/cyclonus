package generator

import (
	"github.com/pkg/errors"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Netpol helps us to avoid the To/From Ingress/Egress dance
type Netpol struct {
	Name         string
	Namespace    string
	PodSelector  metav1.LabelSelector
	IsIngress    bool
	IsEgress     bool
	IngressRules []*Rule
	EgressRules  []*Rule
}

func (n *Netpol) NetworkPolicy() *NetworkPolicy {
	var types []PolicyType
	if n.IsIngress || len(n.IngressRules) > 0 {
		types = append(types, PolicyTypeIngress)
	}
	if n.IsEgress || len(n.EgressRules) > 0 {
		types = append(types, PolicyTypeEgress)
	}
	if len(types) == 0 {
		panic(errors.Errorf("cannot have 0 policy types"))
	}
	var ingress []NetworkPolicyIngressRule
	for _, rule := range n.IngressRules {
		ingress = append(ingress, rule.Ingress())
	}
	var egress []NetworkPolicyEgressRule
	for _, rule := range n.EgressRules {
		egress = append(egress, rule.Egress())
	}
	return &NetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NetworkPolicy",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      n.Name,
			Namespace: n.Namespace,
		},
		Spec: NetworkPolicySpec{
			PodSelector: n.PodSelector,
			Ingress:     ingress,
			Egress:      egress,
			PolicyTypes: types,
		},
	}
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
