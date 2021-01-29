package generator

import (
	v1 "k8s.io/api/core/v1"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	port81 = intstr.FromInt(81)
)

type UpstreamE2EGenerator struct{}

func (u *UpstreamE2EGenerator) GenerateTestCases() []*TestCase {
	return []*TestCase{
		NewTestCase("should support a 'default-deny-ingress' policy", 80, v1.ProtocolTCP, []*Action{CreatePolicy(&NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deny-ingress",
				Namespace: "x",
			},
			Spec: NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress:     []NetworkPolicyIngressRule{},
			},
		})}),
		NewTestCase("should support a 'default-deny-all' policy", 80, v1.ProtocolTCP, []*Action{CreatePolicy(&NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deny-all-tcp-allow-dns",
				Namespace: "x",
			},
			Spec: NetworkPolicySpec{
				PolicyTypes: []PolicyType{PolicyTypeEgress, PolicyTypeIngress},
				PodSelector: metav1.LabelSelector{},
				Ingress:     []NetworkPolicyIngressRule{},
				Egress:      []NetworkPolicyEgressRule{AllowDNSRule.Egress()},
			},
		})}),
		NewTestCase("should enforce policy based on Multiple PodSelectors and NamespaceSelectors", 80, v1.ProtocolTCP, []*Action{CreatePolicy(&NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "allow-ns-y-z-pod-b-c",
				Namespace: "x",
			},
			Spec: NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{"pod": "a"},
				},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress: []NetworkPolicyIngressRule{{
					From: []NetworkPolicyPeer{{
						NamespaceSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{{
								Key:      "ns",
								Operator: metav1.LabelSelectorOpNotIn,
								Values:   []string{"x"},
							}},
						},
						PodSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{{
								Key:      "pod",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"b", "c"},
							}},
						},
					}},
				}},
			},
		})}),
		&TestCase{
			Description: "should enforce multiple, stacked policies with overlapping podSelectors [Feature:NetworkPolicy]",
			Steps: []*TestStep{
				{
					Port:     81,
					Protocol: tcp,
					Actions: []*Action{
						CreatePolicy(&NetworkPolicy{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "allow-client-a-via-ns-selector-81",
								Namespace: "x",
							},
							Spec: NetworkPolicySpec{
								PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{"pod": "a"}},
								Ingress: []NetworkPolicyIngressRule{{
									From: []NetworkPolicyPeer{{
										NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"ns": "y"}},
									}},
									Ports: []NetworkPolicyPort{{Port: &port81, Protocol: &tcp}},
								}},
								PolicyTypes: []PolicyType{PolicyTypeIngress},
							},
						}),
					},
				},
				{
					Port:     80,
					Protocol: tcp,
					Actions:  []*Action{},
				},
				{
					Port:     80,
					Protocol: tcp,
					Actions: []*Action{
						CreatePolicy(&NetworkPolicy{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "allow-client-a-via-ns-selector-80",
								Namespace: "x",
							},
							Spec: NetworkPolicySpec{
								PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{"pod": "a"}},
								Ingress: []NetworkPolicyIngressRule{{
									From: []NetworkPolicyPeer{{
										NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"ns": "y"}},
									}},
									Ports: []NetworkPolicyPort{{Port: &port80, Protocol: &tcp}},
								}},
								PolicyTypes: []PolicyType{PolicyTypeIngress},
							},
						}),
					},
				},
			},
		},
	}
}
