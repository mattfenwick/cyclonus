package matcher

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/netpol/utils"
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
	allowAllOnSCTP := BuildNetworkPolicy(kubePolicy)

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
				PortProtocol: &PortProtocol{Protocol: v1.ProtocolTCP, Port: port103},
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
				PortProtocol: &PortProtocol{Protocol: v1.ProtocolSCTP, Port: port103},
			})
			Expect(sctpAllowed.IsAllowed()).To(BeTrue())
		})
	})

	Describe("Allowing a protocol should implicitly deny other protocols from ips", func() {
		bs, err := json.MarshalIndent(allowAllOnSCTP, "", "  ")
		Expect(err).To(BeNil())
		fmt.Printf("%s\n\n", bs)
		fmt.Printf("%s\n\n", Explain(allowAllOnSCTP))

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
				PortProtocol: &PortProtocol{Protocol: v1.ProtocolTCP, Port: port103},
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
				PortProtocol: &PortProtocol{Protocol: v1.ProtocolSCTP, Port: port103},
			})
			Expect(sctpAllowed.IsAllowed()).To(BeTrue())
		})
	})
}
