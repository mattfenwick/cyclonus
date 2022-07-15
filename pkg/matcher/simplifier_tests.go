package matcher

import (
	"github.com/mattfenwick/cyclonus/pkg/kube/netpol"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func RunSimplifierTests() {
	Describe("Simplifier", func() {
		all := &AllPeersMatcher{}
		allOnTCP80 := &PortsForAllPeersMatcher{Port: &SpecificPortMatcher{Ports: []*PortProtocolMatcher{{Port: &port80, Protocol: tcp}}}}
		allOnTCP103 := &PortsForAllPeersMatcher{Port: &SpecificPortMatcher{Ports: []*PortProtocolMatcher{{Port: &port103, Protocol: tcp}}}}
		allOnTCP80_103 := &PortsForAllPeersMatcher{Port: &SpecificPortMatcher{Ports: []*PortProtocolMatcher{
			{Port: &port80, Protocol: tcp},
			{Port: &port103, Protocol: tcp},
		}}}
		ip := &IPPeerMatcher{
			IPBlock: netpol.IPBlock_10_0_0_1_24,
			Port:    &AllPortMatcher{},
		}
		allPodsAllPorts := &PodPeerMatcher{
			Namespace: &AllNamespaceMatcher{},
			Pod:       &AllPodMatcher{},
			Port:      &AllPortMatcher{},
		}
		allPodsTCP103 := &PodPeerMatcher{
			Namespace: &AllNamespaceMatcher{},
			Pod:       &AllPodMatcher{},
			Port:      &SpecificPortMatcher{Ports: []*PortProtocolMatcher{{Port: &port103, Protocol: tcp}}},
		}

		It("should combine Peer matchers correctly", func() {
			Expect(Simplify([]PeerMatcher{})).To(BeNil())

			Expect(Simplify([]PeerMatcher{all, allOnTCP80, ip, allPodsAllPorts, allPodsTCP103})).To(Equal([]PeerMatcher{all}))

			Expect(Simplify([]PeerMatcher{allPodsAllPorts})).To(Equal([]PeerMatcher{allPodsAllPorts}))
			Expect(Simplify([]PeerMatcher{allOnTCP80})).To(Equal([]PeerMatcher{allOnTCP80}))
			Expect(Simplify([]PeerMatcher{allOnTCP80, allPodsAllPorts})).To(Equal([]PeerMatcher{allOnTCP80, allPodsAllPorts}))
		})

		It("simplifyPodMatchers", func() {
			Expect(simplifyPodMatchers([]*PodPeerMatcher{allPodsAllPorts})).To(Equal([]*PodPeerMatcher{allPodsAllPorts}))
		})

		It("simplifyIPsAndPodsIntoAlls", func() {
			Expect(simplifyIPsAndPodsIntoAlls(nil, nil, nil)).To(BeNil())
			Expect(simplifyIPsAndPodsIntoAlls(allOnTCP80, nil, nil)).To(BeNil())

			ips1, pods1 := simplifyIPsAndPodsIntoAlls(allOnTCP80, nil, []*PodPeerMatcher{allPodsAllPorts})
			Expect(ips1).To(BeNil())
			Expect(pods1).To(Equal([]*PodPeerMatcher{allPodsAllPorts}))
		})

		It("SubtractPortMatchers", func() {
			isEmpty, remaining1 := SubtractPortMatchers(allOnTCP80.Port, allPodsAllPorts.Port)
			Expect(isEmpty).To(BeTrue())
			Expect(remaining1).To(BeNil())

			isEmpty, remaining2 := SubtractPortMatchers(allPodsAllPorts.Port, allOnTCP80.Port)
			Expect(isEmpty).To(BeFalse())
			Expect(remaining2).To(Equal(allPodsAllPorts.Port))
		})

		It("should handle simplifyPortsForAllPeers correctly", func() {
			Expect(simplifyPortsForAllPeers([]*PortsForAllPeersMatcher{})).To(BeNil())
			Expect(simplifyPortsForAllPeers([]*PortsForAllPeersMatcher{allOnTCP80})).To(Equal(allOnTCP80))
			Expect(simplifyPortsForAllPeers([]*PortsForAllPeersMatcher{allOnTCP80, allOnTCP103})).To(Equal(allOnTCP80_103))
			Expect(simplifyPortsForAllPeers([]*PortsForAllPeersMatcher{allOnTCP103, allOnTCP80})).To(Equal(allOnTCP80_103))
		})

		It("Should avoid mixing different matchers -- port range", func() {
			dns := &PortsForAllPeersMatcher{Port: &SpecificPortMatcher{Ports: []*PortProtocolMatcher{{
				Port:     &port53,
				Protocol: udp,
			}}}}
			somePod := &PodPeerMatcher{
				Namespace: &AllNamespaceMatcher{},
				Pod: &LabelSelectorPodMatcher{Selector: metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "x"},
				}},
				Port: &SpecificPortMatcher{PortRanges: []*PortRangeMatcher{{
					From:     80,
					To:       103,
					Protocol: tcp,
				}}},
			}
			Expect(Simplify([]PeerMatcher{dns, somePod})).To(Equal([]PeerMatcher{dns, somePod}))
		})

		It("Should avoid mixing different matchers -- single port", func() {
			dns := &PortsForAllPeersMatcher{Port: &SpecificPortMatcher{Ports: []*PortProtocolMatcher{{
				Port:     &port53,
				Protocol: udp,
			}}}}
			somePod := &PodPeerMatcher{
				Namespace: &AllNamespaceMatcher{},
				Pod: &LabelSelectorPodMatcher{Selector: metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "x"},
				}},
				Port: &SpecificPortMatcher{Ports: []*PortProtocolMatcher{{
					Port:     &port80,
					Protocol: tcp,
				}}},
			}
			Expect(Simplify([]PeerMatcher{dns, somePod})).To(Equal([]PeerMatcher{dns, somePod}))
		})
	})

	Describe("Port Simplifier", func() {
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

			Expect(CombinePortMatchers(allMatcher, allPortsOnSctpMatcher)).To(Equal(allMatcher))
			Expect(CombinePortMatchers(allMatcher, port99OnUdpMatcher)).To(Equal(allMatcher))
			Expect(CombinePortMatchers(allPortsOnSctpMatcher, allMatcher)).To(Equal(allMatcher))
			Expect(CombinePortMatchers(port99OnUdpMatcher, allMatcher)).To(Equal(allMatcher))

			Expect(CombinePortMatchers(allPortsOnSctpMatcher, port99OnUdpMatcher)).To(Equal(combinedMatcher))
			Expect(CombinePortMatchers(port99OnUdpMatcher, allPortsOnSctpMatcher)).To(Equal(combinedMatcher))
		})
	})
}
