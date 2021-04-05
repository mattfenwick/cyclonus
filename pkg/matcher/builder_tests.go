package matcher

import (
	"github.com/mattfenwick/cyclonus/pkg/kube/netpol"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	udp  = v1.ProtocolUDP
	sctp = v1.ProtocolSCTP
	tcp  = v1.ProtocolTCP

	port53  = intstr.FromInt(53)
	port80  = intstr.FromInt(80)
	port103 = intstr.FromInt(103)
)

func RunBuilderTests() {
	Describe("BuildTarget: Allow none -- nil egress/ingress", func() {
		It("allow-no-ingress", func() {
			ingress, egress := BuildTarget(netpol.AllowNoIngress)

			Expect(ingress).ToNot(BeNil())
			Expect(ingress.Peers).To(BeNil())

			Expect(egress).To(BeNil())
		})

		It("allow-no-egress", func() {
			ingress, egress := BuildTarget(netpol.AllowNoEgress)

			Expect(egress).ToNot(BeNil())
			Expect(egress.Peers).To(BeNil())

			Expect(ingress).To(BeNil())
		})

		It("allow-neither", func() {
			ingress, egress := BuildTarget(netpol.AllowNoIngressAllowNoEgress)

			Expect(egress).ToNot(BeNil())
			Expect(egress.Peers).To(BeNil())

			Expect(ingress).ToNot(BeNil())
			Expect(ingress.Peers).To(BeNil())
		})
	})

	Describe("BuildTarget: missing namespace gets treated as default namespace", func() {
		It("missing namespace", func() {
			ingress, egress := BuildTarget(&networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "abc",
				},
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{},
					Ingress:     []networkingv1.NetworkPolicyIngressRule{},
					PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress},
				}})

			Expect(ingress.Namespace).To(Equal("default"))
			Expect(egress.Namespace).To(Equal("default"))
		})
	})

	Describe("BuildTarget: Allow none -- empty ingress/egress", func() {
		It("allow-no-ingress", func() {
			ingress, egress := BuildTarget(netpol.AllowNoIngress_EmptyIngress)

			Expect(ingress).ToNot(BeNil())
			Expect(ingress.Peers).To(BeNil())

			Expect(egress).To(BeNil())
		})

		It("allow-no-egress", func() {
			ingress, egress := BuildTarget(netpol.AllowNoEgress_EmptyEgress)

			Expect(egress).ToNot(BeNil())
			Expect(egress.Peers).To(BeNil())

			Expect(ingress).To(BeNil())
		})

		It("allow-neither", func() {
			ingress, egress := BuildTarget(netpol.AllowNoIngressAllowNoEgress_EmptyEgressEmptyIngress)

			Expect(egress).ToNot(BeNil())
			Expect(egress.Peers).To(BeNil())

			Expect(ingress).ToNot(BeNil())
			Expect(ingress.Peers).To(BeNil())
		})
	})

	Describe("BuildTarget: Allow all", func() {
		It("allow-all-ingress", func() {
			ingress, egress := BuildTarget(netpol.AllowAllIngress)

			Expect(egress).To(BeNil())
			Expect(ingress.Peers).To(Equal([]PeerMatcher{AllPeersPorts}))
		})

		It("allow-all-egress", func() {
			ingress, egress := BuildTarget(netpol.AllowAllEgress)

			Expect(egress.Peers).To(Equal([]PeerMatcher{AllPeersPorts}))
			Expect(ingress).To(BeNil())
		})

		It("allow-all-both", func() {
			ingress, egress := BuildTarget(netpol.AllowAllIngressAllowAllEgress)

			Expect(egress.Peers).To(Equal([]PeerMatcher{AllPeersPorts}))
			Expect(ingress.Peers).To(Equal([]PeerMatcher{AllPeersPorts}))
		})
	})

	// TODO target: combine, ??? etc.

	Describe("PeerMatcher from slice of ingress/egress rules", func() {
		It("allows no ingress from an empty slice of ingress rules", func() {
			peers := BuildIngressMatcher("abc", []networkingv1.NetworkPolicyIngressRule{})
			Expect(peers).To(BeNil())
		})

		It("allows no egress from an empty slice of egress rules", func() {
			peers := BuildEgressMatcher("abc", []networkingv1.NetworkPolicyEgressRule{})
			Expect(peers).To(BeNil())
		})

		It("allows all ingress from an ingress containing a single empty rule", func() {
			peers := BuildIngressMatcher("abc", []networkingv1.NetworkPolicyIngressRule{
				{Ports: nil, From: nil},
			})
			Expect(peers).To(Equal([]PeerMatcher{AllPeersPorts}))
		})

		It("allows all egress from an ingress containing a single empty rule", func() {
			peers := BuildEgressMatcher("abc", []networkingv1.NetworkPolicyEgressRule{
				{Ports: nil, To: nil},
			})
			Expect(peers).To(Equal([]PeerMatcher{AllPeersPorts}))
		})

		It("allows to ips in IPBlock range and also to all pods/ips for DNS", func() {
			peers := BuildEgressMatcher("abc", []networkingv1.NetworkPolicyEgressRule{
				{
					Ports: []networkingv1.NetworkPolicyPort{{Port: &port80, Protocol: &tcp}},
					To: []networkingv1.NetworkPolicyPeer{
						{PodSelector: &metav1.LabelSelector{}},
						{IPBlock: netpol.IPBlock_192_168_242_213_24},
					},
				},
				{
					Ports: []networkingv1.NetworkPolicyPort{{Port: &port53, Protocol: &udp}},
				},
			})
			port53UDPMatcher := &SpecificPortMatcher{Ports: []*PortProtocolMatcher{{Port: &port53, Protocol: v1.ProtocolUDP}}}
			port80TCPMatcher := &SpecificPortMatcher{Ports: []*PortProtocolMatcher{{Port: &port80, Protocol: v1.ProtocolTCP}}}
			ip := &IPPeerMatcher{
				IPBlock: netpol.IPBlock_192_168_242_213_24,
				Port:    port80TCPMatcher,
			}
			Expect(peers).To(Equal([]PeerMatcher{
				&PodPeerMatcher{
					Namespace: &ExactNamespaceMatcher{Namespace: "abc"},
					Pod:       &AllPodMatcher{},
					Port:      port80TCPMatcher,
				},
				ip,
				&PortsForAllPeersMatcher{
					Port: port53UDPMatcher,
				}}))

			Expect(ip.Allows(&TrafficPeer{IP: "192.168.242.249"}, 80, "", tcp)).To(Equal(true))
		})
	})

	Describe("PeerMatcher from slice of NetworkPolicyPeer", func() {
		It("allows all source/destination from an empty slice", func() {
			sds := BuildPeerMatcher("abc", []networkingv1.NetworkPolicyPort{}, []networkingv1.NetworkPolicyPeer{})
			Expect(sds).To(Equal([]PeerMatcher{AllPeersPorts}))
		})

		It("allows all ips and all pods over a specific port from an empty peer slice", func() {
			sds := BuildPeerMatcher("abc", []networkingv1.NetworkPolicyPort{{
				Protocol: &sctp,
				Port:     &port103,
			}}, []networkingv1.NetworkPolicyPeer{})
			portMatcher := &SpecificPortMatcher{Ports: []*PortProtocolMatcher{
				{Port: &port103, Protocol: v1.ProtocolSCTP},
			}}
			matcher := &PortsForAllPeersMatcher{Port: portMatcher}
			Expect(sds).To(Equal([]PeerMatcher{matcher}))
		})

		It("allows ips, but no pods from a single IPBlock", func() {
			peers := BuildPeerMatcher("abc", []networkingv1.NetworkPolicyPort{}, []networkingv1.NetworkPolicyPeer{
				{IPBlock: netpol.IPBlock_10_0_0_1_24},
			})
			ip := &IPPeerMatcher{
				IPBlock: netpol.IPBlock_10_0_0_1_24,
				Port:    &AllPortMatcher{},
			}
			Expect(peers).To(Equal([]PeerMatcher{ip}))
		})

		It("allows all ns/pods/ports, but no ips from a single peer with empty pod/ns selectors", func() {
			peers := BuildPeerMatcher("abc", []networkingv1.NetworkPolicyPort{}, []networkingv1.NetworkPolicyPeer{
				{
					PodSelector:       netpol.SelectorEmpty,
					NamespaceSelector: netpol.SelectorEmpty,
				},
			})
			Expect(peers).To(Equal([]PeerMatcher{
				&PodPeerMatcher{Namespace: &AllNamespaceMatcher{}, Pod: &AllPodMatcher{}, Port: &AllPortMatcher{}}}))
		})

		It("allows ns/pods, but no ips from a single namespace/pod", func() {
			peers := BuildPeerMatcher("abc", []networkingv1.NetworkPolicyPort{}, []networkingv1.NetworkPolicyPeer{
				{PodSelector: netpol.SelectorEmpty},
			})
			matcher := &PodPeerMatcher{
				Namespace: &ExactNamespaceMatcher{Namespace: "abc"},
				Pod:       &AllPodMatcher{},
				Port:      &AllPortMatcher{},
			}
			Expect(peers).To(Equal([]PeerMatcher{matcher}))
		})
	})

	Describe("Namespace/Pod/IPBlock matcher from NetworkPolicyPeer", func() {
		It("allow all pods in policy namespace", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(netpol.Namespace, netpol.AllowAllPodsInPolicyNamespacePeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&ExactNamespaceMatcher{Namespace: netpol.Namespace}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow all pods in all namespaces", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(netpol.Namespace, netpol.AllowAllPodsInAllNamespacesPeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&AllNamespaceMatcher{}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow all pods in matching namespace", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(netpol.Namespace, netpol.AllowAllPodsInMatchingNamespacesPeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&LabelSelectorNamespaceMatcher{Selector: *netpol.SelectorAB}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow all pods in policy namespace -- empty pod selector", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(netpol.Namespace, netpol.AllowAllPodsInPolicyNamespacePeer_EmptyPodSelector)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&ExactNamespaceMatcher{Namespace: netpol.Namespace}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow all pods in all namespaces -- empty pod selector", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(netpol.Namespace, netpol.AllowAllPodsInAllNamespacesPeer_EmptyPodSelector)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&AllNamespaceMatcher{}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow all pods in matching namespace -- empty pod selector", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(netpol.Namespace, netpol.AllowAllPodsInMatchingNamespacesPeer_EmptyPodSelector)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&LabelSelectorNamespaceMatcher{Selector: *netpol.SelectorAB}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow matching pods in policy namespace", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(netpol.Namespace, netpol.AllowMatchingPodsInPolicyNamespacePeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&ExactNamespaceMatcher{Namespace: netpol.Namespace}))
			Expect(pod).To(Equal(&LabelSelectorPodMatcher{Selector: *netpol.SelectorCD}))
		})

		It("allow matching pods in all namespaces", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(netpol.Namespace, netpol.AllowMatchingPodsInAllNamespacesPeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&AllNamespaceMatcher{}))
			Expect(pod).To(Equal(&LabelSelectorPodMatcher{Selector: *netpol.SelectorEF}))
		})

		It("allow matching pods in matching namespace", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(netpol.Namespace, netpol.AllowMatchingPodsInMatchingNamespacesPeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&LabelSelectorNamespaceMatcher{Selector: *netpol.SelectorAB}))
			Expect(pod).To(Equal(&LabelSelectorPodMatcher{Selector: *netpol.SelectorGH}))
		})

		It("allow ipblock", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(netpol.Namespace, netpol.AllowIPBlockPeer)
			Expect(ip).To(Equal(&IPPeerMatcher{
				IPBlock: netpol.IPBlock_10_0_0_1_24,
				Port:    nil,
			}))
			Expect(ns).To(BeNil())
			Expect(pod).To(BeNil())
		})
	})

	Describe("Port from NetworkPolicyPort", func() {
		It("allows all ports and all protocols from an empty slice", func() {
			pm := BuildPortMatcher([]networkingv1.NetworkPolicyPort{})
			Expect(pm).To(Equal(&AllPortMatcher{}))
		})

		It("allow all ports on protocol", func() {
			pm := BuildPortMatcher([]networkingv1.NetworkPolicyPort{netpol.AllowAllPortsOnProtocol})
			Expect(pm).To(Equal(&SpecificPortMatcher{Ports: []*PortProtocolMatcher{{Port: nil, Protocol: v1.ProtocolSCTP}}}))
		})

		It("allow numbered port on protocol", func() {
			portNumber := intstr.FromInt(9001)
			pm := BuildPortMatcher([]networkingv1.NetworkPolicyPort{netpol.AllowNumberedPortOnProtocol})
			Expect(pm).To(Equal(&SpecificPortMatcher{Ports: []*PortProtocolMatcher{{
				Protocol: v1.ProtocolTCP,
				Port:     &portNumber,
			}}}))
		})

		It("allow named port on protocol", func() {
			portName := intstr.FromString("hello")
			pm := BuildPortMatcher([]networkingv1.NetworkPolicyPort{netpol.AllowNamedPortOnProtocol})
			Expect(pm).To(Equal(&SpecificPortMatcher{Ports: []*PortProtocolMatcher{{
				Protocol: v1.ProtocolUDP,
				Port:     &portName,
			}}}))
		})
	})
}
