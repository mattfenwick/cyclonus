package matcher

import (
	"github.com/mattfenwick/cyclonus/pkg/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/yaml"
)

func RunPolicyTests() {
	allowAllOnSCTPSerialized := `
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: policy-207
  namespace: x
spec:
  ingress:
  - ports:
    - protocol: SCTP
  podSelector: {}
  policyTypes:
  - Ingress`
	var kubePolicy *networkingv1.NetworkPolicy
	err := yaml.Unmarshal([]byte(allowAllOnSCTPSerialized), &kubePolicy)
	utils.DoOrDie(err)
	allowAllOnSCTP := BuildNetworkPolicies(true, []*networkingv1.NetworkPolicy{kubePolicy})

	Describe("Allowing a protocol should implicitly deny other protocols from pods", func() {
		It("should not allow TCP", func() {
			tcpAllowed := allowAllOnSCTP.IsTrafficAllowed(&Traffic{
				Source: &TrafficPeer{
					Internal: &InternalPeer{
						PodLabels:       nil,
						NamespaceLabels: nil,
						Namespace:       "y",
					},
					IP: "1.2.3.4",
				},
				Destination: &TrafficPeer{
					Internal: &InternalPeer{
						PodLabels:       nil,
						NamespaceLabels: nil,
						Namespace:       "x",
					},
					IP: "1.2.3.5",
				},
				ResolvedPort: 103,
				//ResolvedPortName: "port-hello",
				Protocol: v1.ProtocolTCP,
			})
			Expect(tcpAllowed.IsAllowed()).To(BeFalse())
		})

		It("should allow SCTP", func() {
			sctpAllowed := allowAllOnSCTP.IsTrafficAllowed(&Traffic{
				Source: &TrafficPeer{
					Internal: &InternalPeer{
						PodLabels:       nil,
						NamespaceLabels: nil,
						Namespace:       "y",
					},
					IP: "1.2.3.4",
				},
				Destination: &TrafficPeer{
					Internal: &InternalPeer{
						PodLabels:       nil,
						NamespaceLabels: nil,
						Namespace:       "x",
					},
					IP: "1.2.3.5",
				},
				ResolvedPort: 103,
				//ResolvedPortName: "port-hello",
				Protocol: v1.ProtocolSCTP,
			})
			Expect(sctpAllowed.IsAllowed()).To(BeTrue())
		})
	})

	Describe("Allowing a protocol should implicitly deny other protocols from ips", func() {
		It("should not allow TCP", func() {
			tcpAllowed := allowAllOnSCTP.IsTrafficAllowed(&Traffic{
				Source: &TrafficPeer{
					Internal: nil,
					IP:       "1.2.3.4",
				},
				Destination: &TrafficPeer{
					Internal: &InternalPeer{
						PodLabels:       nil,
						NamespaceLabels: nil,
						Namespace:       "x",
					},
					IP: "1.2.3.5",
				},
				ResolvedPort: 103,
				//ResolvedPortName: "port-hello",
				Protocol: v1.ProtocolTCP,
			})
			Expect(tcpAllowed.IsAllowed()).To(BeFalse())
		})

		It("should allow SCTP", func() {
			sctpAllowed := allowAllOnSCTP.IsTrafficAllowed(&Traffic{
				Source: &TrafficPeer{
					Internal: nil,
					IP:       "1.2.3.4",
				},
				Destination: &TrafficPeer{
					Internal: &InternalPeer{
						PodLabels:       nil,
						NamespaceLabels: nil,
						Namespace:       "x",
					},
					IP: "1.2.3.5",
				},
				ResolvedPort: 103,
				//ResolvedPortName: "port-hello",
				Protocol: v1.ProtocolSCTP,
			})
			Expect(sctpAllowed.IsAllowed()).To(BeTrue())
		})
	})

	Describe("Policy allowing egress to ips", func() {
		allowAllOnSCTPSerialized := `
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  creationTimestamp: null
  name: vary-egress-37-0-0-0-19
  namespace: x
spec:
  egress:
  - ports:
    - port: 80
      protocol: TCP
    to:
    - podSelector: {}
    - ipBlock:
        cidr: 192.168.242.213/24
  - ports:
    - port: 53
      protocol: UDP
  podSelector:
    matchLabels:
      pod: a
  policyTypes:
  - Egress`
		var kubePolicy *networkingv1.NetworkPolicy
		err := yaml.Unmarshal([]byte(allowAllOnSCTPSerialized), &kubePolicy)
		utils.DoOrDie(err)
		policy := BuildNetworkPolicies(true, []*networkingv1.NetworkPolicy{kubePolicy})

		It("Should allow ips in cidr", func() {
			Expect(policy.IsTrafficAllowed(&Traffic{
				Source: &TrafficPeer{
					Internal: &InternalPeer{
						PodLabels:       map[string]string{"pod": "a"},
						NamespaceLabels: map[string]string{"ns": "x"},
						Namespace:       "x",
					},
					IP: "1.2.3.4",
				},
				Destination: &TrafficPeer{
					Internal: &InternalPeer{
						PodLabels:       map[string]string{"pod": "b"},
						NamespaceLabels: map[string]string{"ns": "y"},
						Namespace:       "y",
					},
					IP: "192.168.242.249",
				},
				ResolvedPort: 80,
				//ResolvedPortName: "port-hello",
				Protocol: v1.ProtocolTCP,
			}).IsAllowed()).To(BeTrue())
		})
	})

	Describe("Policy allowing ingress to named port", func() {
		policyYaml := `
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: abc
  namespace: x
spec:
  ingress:
  - ports:
    - port: port-hello
      protocol: TCP
  podSelector:
    matchLabels:
      pod: a
  policyTypes:
  - Ingress`
		var kubePolicy *networkingv1.NetworkPolicy
		err := yaml.Unmarshal([]byte(policyYaml), &kubePolicy)
		utils.DoOrDie(err)
		policy := BuildNetworkPolicies(true, []*networkingv1.NetworkPolicy{kubePolicy})

		It("Should allow access to named port", func() {
			Expect(policy.IsTrafficAllowed(&Traffic{
				Source: &TrafficPeer{
					IP: "1.2.3.4",
				},
				Destination: &TrafficPeer{
					Internal: &InternalPeer{
						PodLabels:       map[string]string{"pod": "a"},
						NamespaceLabels: map[string]string{"ns": "x"},
						Namespace:       "x",
					},
					IP: "192.168.242.249",
				},
				//ResolvedPort: 0, // TODO
				ResolvedPortName: "port-hello",
				Protocol:         v1.ProtocolTCP,
			}).IsAllowed()).To(BeTrue())
		})
	})
}
