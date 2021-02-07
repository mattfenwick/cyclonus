package generator

import (
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	v1 "k8s.io/api/core/v1"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	port81 = intstr.FromInt(81)

	allAvailable = &ProbeConfig{AllAvailable: true}
	port80TCP    = &ProbeConfig{PortProtocol: &matcher.PortProtocol{Port: port80, Protocol: v1.ProtocolTCP}}
	port81TCP    = &ProbeConfig{PortProtocol: &matcher.PortProtocol{Port: port81, Protocol: v1.ProtocolTCP}}
)

type UpstreamE2EGenerator struct{}

func (u *UpstreamE2EGenerator) GenerateTestCases() []*TestCase {
	return []*TestCase{
		NewSingleStepTestCase("should support a 'default-deny-ingress' policy",
			allAvailable,
			CreatePolicy(&NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deny-ingress",
					Namespace: "x",
				},
				Spec: NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{},
					PolicyTypes: []PolicyType{PolicyTypeIngress},
					Ingress:     []NetworkPolicyIngressRule{},
				},
			})),

		NewSingleStepTestCase("should support a 'default-deny-all' policy",
			allAvailable,
			CreatePolicy(&NetworkPolicy{
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
			})),

		NewSingleStepTestCase("should enforce policy based on Multiple PodSelectors and NamespaceSelectors",
			allAvailable,
			CreatePolicy(&NetworkPolicy{
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
			})),

		NewTestCase("should enforce multiple, stacked policies with overlapping podSelectors [Feature:NetworkPolicy]",
			NewTestStep(allAvailable,
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
			),
			NewTestStep(allAvailable),
			NewTestStep(allAvailable,
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
					}}))),

		NewTestCase("should support allow-all policy",
			NewTestStep(allAvailable, CreatePolicy(
				&NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "allow-all",
						Namespace: "x",
					},
					Spec: NetworkPolicySpec{
						PodSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{},
						},
						Ingress:     []NetworkPolicyIngressRule{{}},
						PolicyTypes: []PolicyType{PolicyTypeIngress},
					}})),
			NewTestStep(allAvailable)),

		NewTestCase("should allow ingress access on one named port",
			NewTestStep(portServe81TCP, CreatePolicy(
				&NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "allow-all",
						Namespace: "x",
					},
					Spec: NetworkPolicySpec{
						PodSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{},
						},
						Ingress: []NetworkPolicyIngressRule{
							{
								Ports: []NetworkPolicyPort{
									{Port: &intstr.IntOrString{Type: intstr.String, StrVal: "serve-81-tcp"}},
								},
							},
						},
						PolicyTypes: []PolicyType{PolicyTypeIngress},
					},
				})),
			NewTestStep(allAvailable)),

		NewTestCase("should enforce updated policy",
			NewTestStep(allAvailable, CreatePolicy(
				&NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "allow-all-mutate-to-deny-all",
						Namespace: "x",
					},
					Spec: NetworkPolicySpec{
						PodSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{},
						},
						Ingress:     []NetworkPolicyIngressRule{{}},
						PolicyTypes: []PolicyType{PolicyTypeIngress},
					},
				})),
			NewTestStep(allAvailable, UpdatePolicy(&NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "allow-all-mutate-to-deny-all",
					Namespace: "x",
				},
				Spec: NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{},
					},
					Ingress:     []NetworkPolicyIngressRule{},
					PolicyTypes: []PolicyType{PolicyTypeIngress},
				},
			}))),

		NewTestCase("should allow ingress access from updated namespace",
			NewTestStep(allAvailable, CreatePolicy(
				&NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "allow-client-a-via-ns-selector",
						Namespace: "x",
					},
					Spec: NetworkPolicySpec{
						PodSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{"pod": "a"},
						},
						Ingress: []NetworkPolicyIngressRule{{
							From: []NetworkPolicyPeer{{
								NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"ns2": "updated"}}},
							},
						}},
						PolicyTypes: []PolicyType{PolicyTypeIngress},
					},
				})),
			NewTestStep(allAvailable, SetNamespaceLabels("y", map[string]string{"ns": "y", "ns2": "updated"}))),

		NewTestCase("should allow ingress access from updated pod",
			NewTestStep(allAvailable, CreatePolicy(
				&NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "allow-client-a-via-pod-selector",
						Namespace: "x",
					},
					Spec: NetworkPolicySpec{
						PodSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{"pod": "a"},
						},
						Ingress: []NetworkPolicyIngressRule{{
							From: []NetworkPolicyPeer{{
								PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"pod": "b", "pod2": "updated"}},
							}},
						}},
						PolicyTypes: []PolicyType{PolicyTypeIngress},
					},
				})),
			NewTestStep(allAvailable, SetPodLabels("x", "b", map[string]string{"pod": "b", "pod2": "updated"}))),

		NewTestCase("should deny ingress access to updated pod",
			NewTestStep(allAvailable, CreatePolicy(
				&NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "deny-ingress-via-label-selector",
						Namespace: "x",
					},
					Spec: NetworkPolicySpec{
						PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{"target": "isolated"}},
						PolicyTypes: []PolicyType{PolicyTypeIngress},
						Ingress:     []NetworkPolicyIngressRule{},
					},
				})),
			NewTestStep(allAvailable, SetPodLabels("x", "a", map[string]string{"target": "isolated"}))),

		NewTestCase("should work with Ingress, Egress specified together",
			NewTestStep(allAvailable, CreatePolicy(
				&NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "allow-client-a-via-pod-selector",
						Namespace: "x",
					},
					Spec: NetworkPolicySpec{
						PodSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{"pod": "a"},
						},
						Ingress: []NetworkPolicyIngressRule{{
							From: []NetworkPolicyPeer{{
								PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"pod": "b"}},
							}},
						}},
						Egress: []NetworkPolicyEgressRule{{
							Ports: []NetworkPolicyPort{
								{Port: &intstr.IntOrString{Type: intstr.Int, IntVal: 80}},
								{Protocol: &udp, Port: &intstr.IntOrString{Type: intstr.Int, IntVal: 53}},
							}}},
						PolicyTypes: []PolicyType{PolicyTypeIngress, PolicyTypeEgress},
					},
				})),
			NewTestStep(allAvailable)),

		NewTestCase("should support denying of egress traffic on the client side (even if the server explicitly allows this traffic)",
			NewTestStep(allAvailable,
				CreatePolicy(
					&NetworkPolicy{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "allow-to-ns-y-pod-a",
							Namespace: "x",
						},
						Spec: NetworkPolicySpec{
							PodSelector: metav1.LabelSelector{
								MatchLabels: map[string]string{"pod": "a"},
							},
							PolicyTypes: []PolicyType{PolicyTypeEgress},
							Egress: []NetworkPolicyEgressRule{
								{To: []NetworkPolicyPeer{{
									NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"ns": "y"}},
									PodSelector:       &metav1.LabelSelector{MatchLabels: map[string]string{"pod": "a"}},
								}}},
								{Ports: []NetworkPolicyPort{{Protocol: &udp, Port: &intstr.IntOrString{Type: intstr.Int, IntVal: 53}}}},
							},
						},
					}),
				CreatePolicy(
					&NetworkPolicy{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "allow-from-xa-on-ya-match-selector",
							Namespace: "y",
						},
						Spec: NetworkPolicySpec{
							PodSelector: metav1.LabelSelector{
								MatchLabels: map[string]string{"pod": "a"},
							},
							Ingress: []NetworkPolicyIngressRule{{
								From: []NetworkPolicyPeer{{
									NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"ns": "x"}},
									PodSelector:       &metav1.LabelSelector{MatchLabels: map[string]string{"pod": "a"}},
								}},
							}},
							PolicyTypes: []PolicyType{PolicyTypeIngress},
						},
					}),
				CreatePolicy(
					&NetworkPolicy{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "allow-from-xa-on-yb-match-selector",
							Namespace: "y",
						},
						Spec: NetworkPolicySpec{
							PodSelector: metav1.LabelSelector{
								MatchLabels: map[string]string{"pod": "b"},
							},
							Ingress: []NetworkPolicyIngressRule{{
								From: []NetworkPolicyPeer{{
									NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"ns": "x"}},
									PodSelector:       &metav1.LabelSelector{MatchLabels: map[string]string{"pod": "a"}},
								}},
							}},
							PolicyTypes: []PolicyType{PolicyTypeIngress},
						},
					}))),

		NewTestCase("should stop enforcing policies after they are deleted",
			NewTestStep(allAvailable, CreatePolicy(
				&NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "deny-all",
						Namespace: "x",
					},
					Spec: NetworkPolicySpec{
						PodSelector: metav1.LabelSelector{},
						PolicyTypes: []PolicyType{PolicyTypeIngress, PolicyTypeEgress},
						Ingress:     []NetworkPolicyIngressRule{},
						Egress:      []NetworkPolicyEgressRule{},
					},
				})),
			NewTestStep(allAvailable, DeletePolicy("x", "deny-all"))),
	}
}
