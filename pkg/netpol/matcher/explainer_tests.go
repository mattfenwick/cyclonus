package matcher

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	udp     = v1.ProtocolUDP
	port103 = intstr.FromInt(103)
	sctp    = v1.ProtocolSCTP
)

func RunExplainerTests() {
	Describe("Explainer", func() {
		It("Complicated ingress", func() {
			complicatedNetpol := &networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "complicated-netpol",
					Namespace: "test-ns",
				},
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{
						MatchLabels:      map[string]string{"pod": "a"},
						MatchExpressions: nil,
					},
					Ingress: []networkingv1.NetworkPolicyIngressRule{
						{
							From: []networkingv1.NetworkPolicyPeer{
								{
									PodSelector:       nil,
									NamespaceSelector: nil,
									IPBlock: &networkingv1.IPBlock{
										CIDR:   "1.2.3.4/24",
										Except: []string{"1.2.3.8"},
									},
								},
							},
						},
						{
							Ports: []networkingv1.NetworkPolicyPort{
								{
									Protocol: &udp,
									Port:     &port103,
								},
								{
									Protocol: &sctp,
									Port:     nil,
								},
							},
							From: []networkingv1.NetworkPolicyPeer{
								{
									PodSelector: &metav1.LabelSelector{
										MatchLabels:      map[string]string{"pod": "b", "stuff": "c"},
										MatchExpressions: nil,
									},
									NamespaceSelector: nil,
									IPBlock:           nil,
								},
								{
									PodSelector: &metav1.LabelSelector{
										MatchLabels:      map[string]string{"pod": "b", "stuff": "c"},
										MatchExpressions: nil,
									},
									NamespaceSelector: &metav1.LabelSelector{
										MatchLabels:      map[string]string{"ns": "y", "other": "z"},
										MatchExpressions: nil,
									},
									IPBlock: nil,
								},
							},
						},
					},
					Egress:      nil,
					PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
				},
			}
			policies := BuildNetworkPolicies([]*networkingv1.NetworkPolicy{complicatedNetpol})
			explanation := Explain(policies)
			fmt.Printf("\n%s\n", explanation)
			expected := `{"Namespace": "test-ns", "PodSelector": ["MatchLabels",["pod: a"],"MatchExpression",null]}
  source rules:
    test-ns/complicated-netpol
  ingress:
    IPBlock: cidr 1.2.3.4/24, except [1.2.3.8]
      all ports all protocols
    namespace test-ns
    pods matching ["MatchLabels",["pod: b","stuff: c"],"MatchExpression",null]
      port 103 on protocol UDP
      all ports on protocol SCTP
    namespaces matching ["MatchLabels",["ns: y","other: z"],"MatchExpression",null]
    pods matching ["MatchLabels",["pod: b","stuff: c"],"MatchExpression",null]
      port 103 on protocol UDP
      all ports on protocol SCTP
`
			Expect(explanation).To(Equal(expected))
		})
	})
}
