package matcher

import (
	"github.com/mattfenwick/cyclonus/pkg/kube/netpol/examples"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var anySourceDestAndPort = &NamespacePodMatcher{
	Namespace: &AllNamespaceMatcher{},
	Pod:       &AllPodMatcher{},
	Port:      &AllPortMatcher{},
}

var anyTrafficPeer = &AllPeerMatcher{}

func RunCornerCaseTests() {
	Describe("Allow none -- nil egress/ingress", func() {
		It("allow-no-ingress", func() {
			ingress, egress := BuildTarget(examples.AllowNoIngress)

			Expect(ingress.Peer).To(Equal(&NonePeerMatcher{}))
			//Expect(target.Ingress).To(Equal(&PeerMatcher{Matchers: []*NamespacePodMatcher{}}))
			Expect(egress).To(BeNil())
		})

		It("allow-no-egress", func() {
			ingress, egress := BuildTarget(examples.AllowNoEgress)

			Expect(egress.Peer).To(Equal(&NonePeerMatcher{}))
			Expect(ingress).To(BeNil())
		})

		It("allow-neither", func() {
			ingress, egress := BuildTarget(examples.AllowNoIngressAllowNoEgress)

			Expect(ingress.Peer).To(Equal(&NonePeerMatcher{}))
			Expect(egress.Peer).To(Equal(&NonePeerMatcher{}))
		})
	})

	Describe("Allow none -- empty ingress/egress", func() {
		It("allow-no-ingress", func() {
			ingress, egress := BuildTarget(examples.AllowNoIngress_EmptyIngress)

			Expect(ingress.Peer).To(Equal(&NonePeerMatcher{}))
			Expect(egress).To(BeNil())
		})

		It("allow-no-egress", func() {
			ingress, egress := BuildTarget(examples.AllowNoEgress_EmptyEgress)

			Expect(egress.Peer).To(Equal(&NonePeerMatcher{}))
			Expect(ingress).To(BeNil())
		})

		It("allow-neither", func() {
			ingress, egress := BuildTarget(examples.AllowNoIngressAllowNoEgress_EmptyEgressEmptyIngress)

			Expect(ingress.Peer).To(Equal(&NonePeerMatcher{}))
			Expect(egress.Peer).To(Equal(&NonePeerMatcher{}))
		})
	})

	Describe("Allow all", func() {
		It("allow-all-ingress", func() {
			ingress, egress := BuildTarget(examples.AllowAllIngress)

			Expect(egress).To(BeNil())
			Expect(ingress.Peer).To(Equal(anyTrafficPeer))
		})

		It("allow-all-egress", func() {
			ingress, egress := BuildTarget(examples.AllowAllEgress)

			Expect(egress.Peer).To(Equal(anyTrafficPeer))
			Expect(ingress).To(BeNil())
		})

		It("allow-all-both", func() {
			ingress, egress := BuildTarget(examples.AllowAllIngressAllowAllEgress)

			Expect(egress.Peer).To(Equal(anyTrafficPeer))
			Expect(ingress.Peer).To(Equal(anyTrafficPeer))
		})
	})

	Describe("Source/destination from slice of NetworkPolicyPeer", func() {
		It("allows all source/destination from an empty slice", func() {
			sds := BuildPeerPortMatchers("abc", []networkingv1.NetworkPolicyPort{}, []networkingv1.NetworkPolicyPeer{})
			Expect(sds).To(Equal(&AllPeerMatcher{}))
		})
	})

	Describe("Source/destination from NetworkPolicyPeer", func() {
		It("allow all pods in policy namespace", func() {
			ip, ns, pod := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInPolicyNamespacePeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&ExactNamespaceMatcher{Namespace: examples.Namespace}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow all pods in all namespaces", func() {
			ip, ns, pod := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInAllNamespacesPeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&AllNamespaceMatcher{}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow all pods in matching namespace", func() {
			ip, ns, pod := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInMatchingNamespacesPeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&LabelSelectorNamespaceMatcher{Selector: *examples.SelectorAB}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow all pods in policy namespace -- empty pod selector", func() {
			ip, ns, pod := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInPolicyNamespacePeer_EmptyPodSelector)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&ExactNamespaceMatcher{Namespace: examples.Namespace}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow all pods in all namespaces -- empty pod selector", func() {
			ip, ns, pod := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInAllNamespacesPeer_EmptyPodSelector)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&AllNamespaceMatcher{}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow all pods in matching namespace -- empty pod selector", func() {
			ip, ns, pod := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInMatchingNamespacesPeer_EmptyPodSelector)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&LabelSelectorNamespaceMatcher{Selector: *examples.SelectorAB}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow matching pods in policy namespace", func() {
			ip, ns, pod := BuildPeerMatcher(examples.Namespace, examples.AllowMatchingPodsInPolicyNamespacePeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&ExactNamespaceMatcher{Namespace: examples.Namespace}))
			Expect(pod).To(Equal(&LabelSelectorPodMatcher{Selector: *examples.SelectorCD}))
		})

		It("allow matching pods in all namespaces", func() {
			ip, ns, pod := BuildPeerMatcher(examples.Namespace, examples.AllowMatchingPodsInAllNamespacesPeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&AllNamespaceMatcher{}))
			Expect(pod).To(Equal(&LabelSelectorPodMatcher{Selector: *examples.SelectorEF}))
		})

		It("allow matching pods in matching namespace", func() {
			ip, ns, pod := BuildPeerMatcher(examples.Namespace, examples.AllowMatchingPodsInMatchingNamespacesPeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&LabelSelectorNamespaceMatcher{Selector: *examples.SelectorAB}))
			Expect(pod).To(Equal(&LabelSelectorPodMatcher{Selector: *examples.SelectorGH}))
		})

		It("allow ipblock", func() {
			ip, ns, pod := BuildPeerMatcher(examples.Namespace, examples.AllowIPBlockPeer)
			Expect(ip).To(Equal(&IPBlockMatcher{
				IPBlock: &networkingv1.IPBlock{CIDR: "10.0.0.1/24",
					Except: []string{
						"10.0.0.2",
					},
				},
				Port: nil,
			}))
			Expect(ns).To(BeNil())
			Expect(pod).To(BeNil())
		})
	})

	Describe("Port from slice of NetworkPolicyPort", func() {
		It("allows all ports and all protocols from an empty slice", func() {
			pm := BuildPortMatcher([]networkingv1.NetworkPolicyPort{})
			Expect(pm).To(Equal(&AllPortMatcher{}))
		})
	})

	Describe("Port from NetworkPolicyPort", func() {
		It("allow all ports on protocol", func() {
			pm := BuildPortMatcher([]networkingv1.NetworkPolicyPort{examples.AllowAllPortsOnProtocol})
			Expect(pm).To(Equal(&SpecificPortsMatcher{Ports: []*PortProtocolMatcher{{Port: nil, Protocol: v1.ProtocolSCTP}}}))
		})

		It("allow numbered port on protocol", func() {
			portNumber := intstr.FromInt(9001)
			pm := BuildPortMatcher([]networkingv1.NetworkPolicyPort{examples.AllowNumberedPortOnProtocol})
			Expect(pm).To(Equal(&SpecificPortsMatcher{[]*PortProtocolMatcher{{
				Protocol: v1.ProtocolTCP,
				Port:     &portNumber,
			}}}))
		})

		It("allow named port on protocol", func() {
			portName := intstr.FromString("hello")
			pm := BuildPortMatcher([]networkingv1.NetworkPolicyPort{examples.AllowNamedPortOnProtocol})
			Expect(pm).To(Equal(&SpecificPortsMatcher{[]*PortProtocolMatcher{{
				Protocol: v1.ProtocolUDP,
				Port:     &portName,
			}}}))
		})
	})
}
