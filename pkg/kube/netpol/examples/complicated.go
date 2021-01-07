package examples

import (
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func ExampleComplicatedNetworkPolicy() *networkingv1.NetworkPolicy {
	tcp := v1.ProtocolTCP
	port3333 := intstr.FromInt(3333)
	port4444 := intstr.FromInt(4444)
	port5555 := intstr.FromInt(5555)
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "example-namespace",
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					Ports: []networkingv1.NetworkPolicyPort{
						{Protocol: &tcp, Port: &port3333},
						{Protocol: &tcp, Port: &port4444},
						{Protocol: &tcp, Port: &port5555},
					},
					From: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels:      nil,
								MatchExpressions: nil,
							},
							NamespaceSelector: nil,
						},
						{
							PodSelector: nil,
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels:      nil,
								MatchExpressions: nil,
							},
						},
						{
							IPBlock: &networkingv1.IPBlock{
								CIDR:   "10.0.0.0/16",
								Except: []string{"10.0.0.0", "10.0.0.1"},
							},
						},
					},
				},
			},
			Egress: nil,
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
		},
	}
}
