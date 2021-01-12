package netpolgen

import (
	"fmt"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewDefaultGenerator() *Generator {
	return &Generator{
		Ports:      DefaultPorts(),
		Peers:      DefaultPeers(),
		Targets:    DefaultTargets(),
		Namespaces: DefaultNamespaces(),
	}
}

type Generator struct {
	Ports      []NetworkPolicyPort
	Peers      []NetworkPolicyPeer
	Targets    []metav1.LabelSelector
	Namespaces []string
}

func (g *Generator) PortSlices() [][]NetworkPolicyPort {
	// 0 length
	slices := [][]NetworkPolicyPort{noPorts}
	// 1 length
	for _, s := range g.Ports {
		slices = append(slices, []NetworkPolicyPort{s})
	}
	// TODO 2 length
	return slices
}

func (g *Generator) PeerSlices() [][]NetworkPolicyPeer {
	// 0 length
	slices := [][]NetworkPolicyPeer{noPeers}
	// 1 length
	for _, p := range g.Peers {
		slices = append(slices, []NetworkPolicyPeer{p})
	}
	// TODO 2 length
	return slices
}

func (g *Generator) IngressRules() []NetworkPolicyIngressRule {
	var is []NetworkPolicyIngressRule
	for _, ports := range g.PortSlices() {
		for _, peers := range g.PeerSlices() {
			is = append(is, NetworkPolicyIngressRule{
				Ports: ports,
				From:  peers,
			})
		}
	}
	return is
}

func (g *Generator) IngressRuleSlices() [][]NetworkPolicyIngressRule {
	// 0 length
	slices := [][]NetworkPolicyIngressRule{noIngressRules}
	// 1 length
	for _, ingress := range g.IngressRules() {
		slices = append(slices, []NetworkPolicyIngressRule{ingress})
	}
	// TODO 2 length
	return slices
}

func (g *Generator) IngressPolicies() []*NetworkPolicy {
	var policies []*NetworkPolicy
	i := 0
	for _, ingress := range g.IngressRuleSlices() {
		for _, target := range g.Targets {
			for _, ns := range g.Namespaces {
				policies = append(policies, &NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("policy-%d", i),
						Namespace: ns,
					},
					Spec: NetworkPolicySpec{
						PodSelector: target,
						Ingress:     ingress,
						PolicyTypes: []PolicyType{PolicyTypeIngress},
					},
				})
				i++
			}
		}
	}
	return policies
}

func (g *Generator) EgressPolicies() []*NetworkPolicy {
	panic("TODO")
}

func (g *Generator) IngressEgressPolicies() []*NetworkPolicy {
	panic("TODO")
}
