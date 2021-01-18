package matcher

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube/netpol/examples"
	"github.com/mattfenwick/cyclonus/pkg/netpol/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	port53 = intstr.FromInt(53)
	port80 = intstr.FromInt(80)

	tcp = v1.ProtocolTCP
)

func RunBuilderTests() {
	Describe("BuildTarget: Allow none -- nil egress/ingress", func() {
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

	Describe("BuildTarget: Allow none -- empty ingress/egress", func() {
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

	Describe("BuildTarget: Allow all", func() {
		It("allow-all-ingress", func() {
			ingress, egress := BuildTarget(examples.AllowAllIngress)

			Expect(egress).To(BeNil())
			Expect(ingress.Peer).To(Equal(&AllPeerMatcher{}))
		})

		It("allow-all-egress", func() {
			ingress, egress := BuildTarget(examples.AllowAllEgress)

			Expect(egress.Peer).To(Equal(&AllPeerMatcher{}))
			Expect(ingress).To(BeNil())
		})

		It("allow-all-both", func() {
			ingress, egress := BuildTarget(examples.AllowAllIngressAllowAllEgress)

			Expect(egress.Peer).To(Equal(&AllPeerMatcher{}))
			Expect(ingress.Peer).To(Equal(&AllPeerMatcher{}))
		})
	})

	// TODO target: combine, ??? etc.

	Describe("PeerMatcher from slice of ingress/egress rules", func() {
		It("allows no ingress from an empty slice of ingress rules", func() {
			peer := BuildIngressMatcher("abc", []networkingv1.NetworkPolicyIngressRule{})
			Expect(peer).To(Equal(&NonePeerMatcher{}))
		})

		It("allows no egress from an empty slice of egress rules", func() {
			peer := BuildEgressMatcher("abc", []networkingv1.NetworkPolicyEgressRule{})
			Expect(peer).To(Equal(&NonePeerMatcher{}))
		})

		It("allows all ingress from an ingress containing a single empty rule", func() {
			peer := BuildIngressMatcher("abc", []networkingv1.NetworkPolicyIngressRule{
				{
					Ports: nil,
					From:  nil,
				},
			})
			Expect(peer).To(Equal(&AllPeerMatcher{}))
		})

		It("allows all egress from an ingress containing a single empty rule", func() {
			peer := BuildEgressMatcher("abc", []networkingv1.NetworkPolicyEgressRule{
				{
					Ports: nil,
					To:    nil,
				},
			})
			Expect(peer).To(Equal(&AllPeerMatcher{}))
		})

		It("allows to ips in IPBlock range and also to all pods/ips for DNS", func() {
			peer := BuildEgressMatcher("abc", []networkingv1.NetworkPolicyEgressRule{
				{
					Ports: []networkingv1.NetworkPolicyPort{{Port: &port80, Protocol: &tcp}},
					To: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{},
						},
						{
							IPBlock: examples.IPBlock_192_168_242_213_24,
						},
					},
				},
				{
					Ports: []networkingv1.NetworkPolicyPort{{Port: &port53, Protocol: &udp}},
				},
			})
			port53UDPMatcher := &SpecificPortMatcher{Ports: []*PortProtocolMatcher{{Port: &port53, Protocol: v1.ProtocolUDP}}}
			port80TCPMatcher := &SpecificPortMatcher{Ports: []*PortProtocolMatcher{{Port: &port80, Protocol: v1.ProtocolTCP}}}
			ip := &IPBlockMatcher{
				IPBlock: examples.IPBlock_192_168_242_213_24,
				Port:    port80TCPMatcher,
			}
			Expect(peer).To(Equal(&SpecificPeerMatcher{
				IP: NewSpecificIPMatcher(port53UDPMatcher, ip),
				Internal: NewSpecificInternalMatcher(&NamespacePodMatcher{
					Namespace: &ExactNamespaceMatcher{Namespace: "abc"},
					Pod:       &AllPodMatcher{},
					Port:      port80TCPMatcher,
				}, &NamespacePodMatcher{
					Namespace: &AllNamespaceMatcher{},
					Pod:       &AllPodMatcher{},
					Port:      port53UDPMatcher,
				}),
			}))
			Expect(ip.Allows("192.168.242.249", &PortProtocol{Port: port80, Protocol: tcp})).To(Equal(true))
		})
	})

	Describe("PeerMatcher from slice of NetworkPolicyPeer", func() {
		It("allows all source/destination from an empty slice", func() {
			sds := BuildPeerMatcher("abc", []networkingv1.NetworkPolicyPort{}, []networkingv1.NetworkPolicyPeer{})
			Expect(sds).To(Equal(&AllPeerMatcher{}))
		})

		It("allows all ips and all pods over a specific port from an empty peer slice", func() {
			sds := BuildPeerMatcher("abc", []networkingv1.NetworkPolicyPort{{
				Protocol: &sctp,
				Port:     &port103,
			}}, []networkingv1.NetworkPolicyPeer{})
			portMatcher := &SpecificPortMatcher{Ports: []*PortProtocolMatcher{
				{
					Port:     &port103,
					Protocol: v1.ProtocolSCTP,
				},
			}}
			matcher := &NamespacePodMatcher{
				Namespace: &AllNamespaceMatcher{},
				Pod:       &AllPodMatcher{},
				Port:      portMatcher,
			}
			Expect(sds).To(Equal(&SpecificPeerMatcher{
				IP:       NewSpecificIPMatcher(portMatcher),
				Internal: NewSpecificInternalMatcher(matcher),
			}))
		})

		It("allows ips, but no pods from a single IPBlock", func() {
			peer := BuildPeerMatcher("abc", []networkingv1.NetworkPolicyPort{}, []networkingv1.NetworkPolicyPeer{
				{
					PodSelector:       nil,
					NamespaceSelector: nil,
					IPBlock:           examples.IPBlock_10_0_0_1_24,
				},
			})
			ip := &IPBlockMatcher{
				IPBlock: examples.IPBlock_10_0_0_1_24,
				Port:    &AllPortMatcher{},
			}
			Expect(peer).To(Equal(&SpecificPeerMatcher{
				IP:       NewSpecificIPMatcher(&NonePortMatcher{}, ip),
				Internal: &NoneInternalMatcher{},
			}))
		})

		It("allows all ns/pods/ports, but no ips from a single peer with empty pod/ns selectors", func() {
			peer := BuildPeerMatcher("abc", []networkingv1.NetworkPolicyPort{}, []networkingv1.NetworkPolicyPeer{
				{
					PodSelector:       examples.SelectorEmpty,
					NamespaceSelector: examples.SelectorEmpty,
					IPBlock:           nil,
				},
			})
			Expect(peer).To(Equal(&SpecificPeerMatcher{
				IP:       &NoneIPMatcher{},
				Internal: &AllInternalMatcher{},
			}))
		})

		It("allows ns/pods, but no ips from a single namespace/pod", func() {
			peer := BuildPeerMatcher("abc", []networkingv1.NetworkPolicyPort{}, []networkingv1.NetworkPolicyPeer{
				{
					PodSelector:       examples.SelectorEmpty,
					NamespaceSelector: nil,
					IPBlock:           nil,
				},
			})
			matcher := &NamespacePodMatcher{
				Namespace: &ExactNamespaceMatcher{Namespace: "abc"},
				Pod:       &AllPodMatcher{},
				Port:      &AllPortMatcher{},
			}
			Expect(peer).To(Equal(&SpecificPeerMatcher{
				IP: &NoneIPMatcher{},
				Internal: &SpecificInternalMatcher{NamespacePods: map[string]*NamespacePodMatcher{
					matcher.PrimaryKey(): matcher,
				}},
			}))
		})

		// TODO SpecificPeerMatcher
		// TODO SpecificInternalMatcher

		It("should combine Peer matchers correctly", func() {
			all := &AllPeerMatcher{}
			none := &NonePeerMatcher{}
			ip := &IPBlockMatcher{
				IPBlock: examples.IPBlock_10_0_0_1_24,
				Port:    &AllPortMatcher{},
			}
			someIps := &SpecificPeerMatcher{
				IP: &SpecificIPMatcher{IPBlocks: map[string]*IPBlockMatcher{
					ip.PrimaryKey(): ip,
				}},
				Internal: &NoneInternalMatcher{},
			}
			someInternal1 := &SpecificPeerMatcher{
				IP:       &NoneIPMatcher{},
				Internal: &AllInternalMatcher{},
			}
			someInternal2 := &SpecificPeerMatcher{
				IP:       &NoneIPMatcher{},
				Internal: &NoneInternalMatcher{},
			}

			Expect(CombinePeerMatchers(all, all)).To(Equal(all))
			Expect(CombinePeerMatchers(all, none)).To(Equal(all))
			Expect(CombinePeerMatchers(all, someIps)).To(Equal(all))
			Expect(CombinePeerMatchers(all, someInternal1)).To(Equal(all))
			Expect(CombinePeerMatchers(none, all)).To(Equal(all))
			Expect(CombinePeerMatchers(someIps, all)).To(Equal(all))
			Expect(CombinePeerMatchers(someInternal1, all)).To(Equal(all))

			Expect(CombinePeerMatchers(none, none)).To(Equal(none))
			Expect(CombinePeerMatchers(none, someIps)).To(Equal(someIps))
			Expect(CombinePeerMatchers(none, someInternal1)).To(Equal(someInternal1))
			Expect(CombinePeerMatchers(someIps, none)).To(Equal(someIps))
			Expect(CombinePeerMatchers(someInternal1, none)).To(Equal(someInternal1))

			bs, err := json.MarshalIndent([]interface{}{someInternal1, someInternal2}, "", "  ")
			utils.DoOrDie(err)
			fmt.Printf("%s\n\n", bs)

			Expect(CombinePeerMatchers(someInternal1, someInternal2)).To(Equal(someInternal1))
			Expect(CombinePeerMatchers(someInternal2, someInternal1)).To(Equal(someInternal1))
		})

		It("should combine Internal matchers correctly", func() {
			all := &AllInternalMatcher{}
			none := &NoneInternalMatcher{}
			np1 := &NamespacePodMatcher{
				Namespace: &AllNamespaceMatcher{},
				Pod:       &LabelSelectorPodMatcher{Selector: *examples.SelectorAB},
				Port:      &AllPortMatcher{},
			}
			some1 := &SpecificInternalMatcher{NamespacePods: map[string]*NamespacePodMatcher{
				np1.PrimaryKey(): np1,
			}}

			Expect(CombineInternalMatchers(all, all)).To(Equal(all))
			Expect(CombineInternalMatchers(all, none)).To(Equal(all))
			Expect(CombineInternalMatchers(all, some1)).To(Equal(all))
			Expect(CombineInternalMatchers(none, all)).To(Equal(all))
			Expect(CombineInternalMatchers(some1, all)).To(Equal(all))

			Expect(CombineInternalMatchers(none, none)).To(Equal(none))
			Expect(CombineInternalMatchers(none, some1)).To(Equal(some1))
			Expect(CombineInternalMatchers(some1, none)).To(Equal(some1))
		})
	})

	Describe("Namespace/Pod/IPBlock matcher from NetworkPolicyPeer", func() {
		It("allow all pods in policy namespace", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(examples.Namespace, examples.AllowAllPodsInPolicyNamespacePeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&ExactNamespaceMatcher{Namespace: examples.Namespace}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow all pods in all namespaces", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(examples.Namespace, examples.AllowAllPodsInAllNamespacesPeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&AllNamespaceMatcher{}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow all pods in matching namespace", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(examples.Namespace, examples.AllowAllPodsInMatchingNamespacesPeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&LabelSelectorNamespaceMatcher{Selector: *examples.SelectorAB}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow all pods in policy namespace -- empty pod selector", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(examples.Namespace, examples.AllowAllPodsInPolicyNamespacePeer_EmptyPodSelector)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&ExactNamespaceMatcher{Namespace: examples.Namespace}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow all pods in all namespaces -- empty pod selector", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(examples.Namespace, examples.AllowAllPodsInAllNamespacesPeer_EmptyPodSelector)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&AllNamespaceMatcher{}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow all pods in matching namespace -- empty pod selector", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(examples.Namespace, examples.AllowAllPodsInMatchingNamespacesPeer_EmptyPodSelector)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&LabelSelectorNamespaceMatcher{Selector: *examples.SelectorAB}))
			Expect(pod).To(Equal(&AllPodMatcher{}))
		})

		It("allow matching pods in policy namespace", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(examples.Namespace, examples.AllowMatchingPodsInPolicyNamespacePeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&ExactNamespaceMatcher{Namespace: examples.Namespace}))
			Expect(pod).To(Equal(&LabelSelectorPodMatcher{Selector: *examples.SelectorCD}))
		})

		It("allow matching pods in all namespaces", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(examples.Namespace, examples.AllowMatchingPodsInAllNamespacesPeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&AllNamespaceMatcher{}))
			Expect(pod).To(Equal(&LabelSelectorPodMatcher{Selector: *examples.SelectorEF}))
		})

		It("allow matching pods in matching namespace", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(examples.Namespace, examples.AllowMatchingPodsInMatchingNamespacesPeer)
			Expect(ip).To(BeNil())
			Expect(ns).To(Equal(&LabelSelectorNamespaceMatcher{Selector: *examples.SelectorAB}))
			Expect(pod).To(Equal(&LabelSelectorPodMatcher{Selector: *examples.SelectorGH}))
		})

		It("allow ipblock", func() {
			ip, ns, pod := BuildIPBlockNamespacePodMatcher(examples.Namespace, examples.AllowIPBlockPeer)
			Expect(ip).To(Equal(&IPBlockMatcher{
				IPBlock: examples.IPBlock_10_0_0_1_24,
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
			pm := BuildPortMatcher([]networkingv1.NetworkPolicyPort{examples.AllowAllPortsOnProtocol})
			Expect(pm).To(Equal(&SpecificPortMatcher{Ports: []*PortProtocolMatcher{{Port: nil, Protocol: v1.ProtocolSCTP}}}))
		})

		It("allow numbered port on protocol", func() {
			portNumber := intstr.FromInt(9001)
			pm := BuildPortMatcher([]networkingv1.NetworkPolicyPort{examples.AllowNumberedPortOnProtocol})
			Expect(pm).To(Equal(&SpecificPortMatcher{[]*PortProtocolMatcher{{
				Protocol: v1.ProtocolTCP,
				Port:     &portNumber,
			}}}))
		})

		It("allow named port on protocol", func() {
			portName := intstr.FromString("hello")
			pm := BuildPortMatcher([]networkingv1.NetworkPolicyPort{examples.AllowNamedPortOnProtocol})
			Expect(pm).To(Equal(&SpecificPortMatcher{[]*PortProtocolMatcher{{
				Protocol: v1.ProtocolUDP,
				Port:     &portName,
			}}}))
		})

		It("should combine matchers correctly", func() {
			port99 := intstr.FromInt(99)
			allPortsOnSctp := &PortProtocolMatcher{
				Port:     nil,
				Protocol: v1.ProtocolSCTP,
			}
			port99OnUdp := &PortProtocolMatcher{
				Port:     &port99,
				Protocol: v1.ProtocolSCTP,
			}

			allMatcher := &AllPortMatcher{}
			allPortsOnSctpMatcher := &SpecificPortMatcher{Ports: []*PortProtocolMatcher{allPortsOnSctp}}
			port99OnUdpMatcher := &SpecificPortMatcher{Ports: []*PortProtocolMatcher{port99OnUdp}}
			combinedMatcher := &SpecificPortMatcher{Ports: []*PortProtocolMatcher{allPortsOnSctp, port99OnUdp}}
			combined2Matcher := &SpecificPortMatcher{Ports: []*PortProtocolMatcher{port99OnUdp, allPortsOnSctp}}

			Expect(CombinePortMatchers(allMatcher, allPortsOnSctpMatcher)).To(Equal(allMatcher))
			Expect(CombinePortMatchers(allMatcher, port99OnUdpMatcher)).To(Equal(allMatcher))
			Expect(CombinePortMatchers(allPortsOnSctpMatcher, allMatcher)).To(Equal(allMatcher))
			Expect(CombinePortMatchers(port99OnUdpMatcher, allMatcher)).To(Equal(allMatcher))

			Expect(CombinePortMatchers(allPortsOnSctpMatcher, port99OnUdpMatcher)).To(Equal(combinedMatcher))
			Expect(CombinePortMatchers(port99OnUdpMatcher, allPortsOnSctpMatcher)).To(Equal(combined2Matcher))
		})
	})
}
