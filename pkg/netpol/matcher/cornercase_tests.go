package matcher

import (
	"github.com/mattfenwick/cyclonus/pkg/kube/netpol/examples"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	//v1 "k8s.io/api/core/v1"
	//networkingv1 "k8s.io/api/networking/v1"
	//"k8s.io/apimachinery/pkg/util/intstr"
)

//var anySourceDestAndPort = &PodPeerMatcher{
//	Pod:  &AnywherePeerMatcher{},
//	Port: &AllPortsAllProtocolsMatcher{},
//}
//
//var anyTrafficPeer = &PeerPortEdgeMatcher{Matchers: []*PodPeerMatcher{anySourceDestAndPort}}

func RunCornerCaseTests() {
	Describe("Allow none -- nil egress/ingress", func() {
		It("allow-no-ingress", func() {
			ingress, egress := BuildTarget(examples.AllowNoIngress)

			Expect(ingress.Edge).To(Equal(&NoneEdgeMatcher{}))
			//Expect(target.Ingress).To(Equal(&EdgeMatcher{Matchers: []*PodPeerMatcher{}}))
			Expect(egress).To(BeNil())
		})

		It("allow-no-egress", func() {
			ingress, egress := BuildTarget(examples.AllowNoEgress)

			Expect(egress.Edge).To(Equal(&NoneEdgeMatcher{}))
			Expect(ingress).To(BeNil())
		})

		It("allow-neither", func() {
			ingress, egress := BuildTarget(examples.AllowNoIngressAllowNoEgress)

			Expect(ingress.Edge).To(Equal(&NoneEdgeMatcher{}))
			Expect(egress.Edge).To(Equal(&NoneEdgeMatcher{}))
		})
	})

	Describe("Allow none -- empty ingress/egress", func() {
		It("allow-no-ingress", func() {
			ingress, egress := BuildTarget(examples.AllowNoIngress_EmptyIngress)

			Expect(ingress.Edge).To(Equal(&NoneEdgeMatcher{}))
			Expect(egress).To(BeNil())
		})

		It("allow-no-egress", func() {
			ingress, egress := BuildTarget(examples.AllowNoEgress_EmptyEgress)

			Expect(egress.Edge).To(Equal(&NoneEdgeMatcher{}))
			Expect(ingress).To(BeNil())
		})

		It("allow-neither", func() {
			ingress, egress := BuildTarget(examples.AllowNoIngressAllowNoEgress_EmptyEgressEmptyIngress)

			Expect(ingress.Edge).To(Equal(&NoneEdgeMatcher{}))
			Expect(egress.Edge).To(Equal(&NoneEdgeMatcher{}))
		})
	})

	//Describe("Allow all", func() {
	//	It("allow-all-ingress", func() {
	//		ingress, egress := BuildTarget(examples.AllowAllIngress)
	//
	//		Expect(egress).To(BeNil())
	//		Expect(ingress.Edge).To(Equal(anyTrafficPeer))
	//	})
	//
	//	It("allow-all-egress", func() {
	//		ingress, egress := BuildTarget(examples.AllowAllEgress)
	//
	//		Expect(egress.Edge).To(Equal(anyTrafficPeer))
	//		Expect(ingress).To(BeNil())
	//	})
	//
	//	It("allow-all-both", func() {
	//		ingress, egress := BuildTarget(examples.AllowAllIngressAllowAllEgress)
	//
	//		Expect(egress.Edge).To(Equal(anyTrafficPeer))
	//		Expect(ingress.Edge).To(Equal(anyTrafficPeer))
	//	})
	//})

	//Describe("Source/destination from slice of NetworkPolicyPeer", func() {
	//	It("allows all source/destination from an empty slice", func() {
	//		sds := BuildPeerMatchers("abc", []networkingv1.NetworkPolicyPeer{})
	//		Expect(sds).To(Equal([]PodMatcher{&AnywherePeerMatcher{}}))
	//	})
	//})
	//
	//Describe("Source/destination from NetworkPolicyPeer", func() {
	//	It("allow all pods in policy namespace", func() {
	//		sd := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInPolicyNamespacePeer)
	//		Expect(sd).To(Equal(&AllPodsInPolicyNamespacePeerMatcher{Namespace: examples.Namespace}))
	//	})
	//
	//	It("allow all pods in all namespaces", func() {
	//		sd := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInAllNamespacesPeer)
	//		Expect(sd).To(Equal(&AllPodsAllNamespacesPeerMatcher{}))
	//	})
	//
	//	It("allow all pods in matching namespace", func() {
	//		sd := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInMatchingNamespacesPeer)
	//		Expect(sd).To(Equal(&AllPodsInMatchingNamespacesPeerMatcher{NamespaceSelector: *examples.SelectorAB}))
	//	})
	//
	//	It("allow all pods in policy namespace -- empty pod selector", func() {
	//		sd := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInPolicyNamespacePeer_EmptyPodSelector)
	//		Expect(sd).To(Equal(&AllPodsInPolicyNamespacePeerMatcher{Namespace: examples.Namespace}))
	//	})
	//
	//	It("allow all pods in all namespaces -- empty pod selector", func() {
	//		sd := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInAllNamespacesPeer_EmptyPodSelector)
	//		Expect(sd).To(Equal(&AllPodsAllNamespacesPeerMatcher{}))
	//	})
	//
	//	It("allow all pods in matching namespace -- empty pod selector", func() {
	//		sd := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInMatchingNamespacesPeer_EmptyPodSelector)
	//		Expect(sd).To(Equal(&AllPodsInMatchingNamespacesPeerMatcher{NamespaceSelector: *examples.SelectorAB}))
	//	})
	//
	//	It("allow matching pods in policy namespace", func() {
	//		sd := BuildPeerMatcher(examples.Namespace, examples.AllowMatchingPodsInPolicyNamespacePeer)
	//		Expect(sd).To(Equal(&MatchingPodsInPolicyNamespacePeerMatcher{PodSelector: *examples.SelectorCD, Namespace: examples.Namespace}))
	//	})
	//
	//	It("allow matching pods in all namespaces", func() {
	//		sd := BuildPeerMatcher(examples.Namespace, examples.AllowMatchingPodsInAllNamespacesPeer)
	//		Expect(sd).To(Equal(&MatchingPodsInAllNamespacesPeerMatcher{PodSelector: *examples.SelectorEF}))
	//	})
	//
	//	It("allow matching pods in matching namespace", func() {
	//		sd := BuildPeerMatcher(examples.Namespace, examples.AllowMatchingPodsInMatchingNamespacesPeer)
	//		Expect(sd).To(Equal(&MatchingPodsInMatchingNamespacesPeerMatcher{
	//			PodSelector:       *examples.SelectorGH,
	//			NamespaceSelector: *examples.SelectorAB,
	//		}))
	//	})
	//
	//	It("allow ipblock", func() {
	//		sd := BuildPeerMatcher(examples.Namespace, examples.AllowIPBlockPeer)
	//		Expect(sd).To(Equal(&IPBlockPeerMatcher{
	//			&networkingv1.IPBlock{CIDR: "10.0.0.1/24",
	//				Except: []string{
	//					"10.0.0.2",
	//				},
	//			},
	//		}))
	//	})
	//})
	//
	//Describe("Port from slice of NetworkPolicyPort", func() {
	//	It("allows all ports and all protocols from an empty slice", func() {
	//		sds := BuildPortMatchers([]networkingv1.NetworkPolicyPort{})
	//		Expect(sds).To(Equal([]PortMatcher{&AllPortsAllProtocolsMatcher{}}))
	//	})
	//})
	//
	//Describe("Port from NetworkPolicyPort", func() {
	//	It("allow all ports on protocol", func() {
	//		sd := BuildPortMatcher(examples.AllowAllPortsOnProtocol)
	//		Expect(sd).To(Equal(&AllPortsOnProtocolMatcher{Protocol: v1.ProtocolSCTP}))
	//	})
	//
	//	It("allow numbered port on protocol", func() {
	//		sd := BuildPortMatcher(examples.AllowNumberedPortOnProtocol)
	//		Expect(sd).To(Equal(&ExactPortProtocolMatcher{
	//			Protocol: v1.ProtocolTCP,
	//			Port:     intstr.FromInt(9001),
	//		}))
	//	})
	//
	//	It("allow named port on protocol", func() {
	//		sd := BuildPortMatcher(examples.AllowNamedPortOnProtocol)
	//		Expect(sd).To(Equal(&ExactPortProtocolMatcher{
	//			Protocol: v1.ProtocolUDP,
	//			Port:     intstr.FromString("hello"),
	//		}))
	//	})
	//})
}
