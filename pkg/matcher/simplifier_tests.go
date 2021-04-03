package matcher

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func RunSimplifierTests() {
	Describe("Simplifier", func() {
		It("should combine Peer matchers correctly", func() {
			Expect(nil).ToNot(BeNil())
			//all := &PortsForAllPeersMatcher{}
			//ip := &IPPeerMatcher{
			//	IPBlock: netpol.IPBlock_10_0_0_1_24,
			//	Port:    &AllPortMatcher{},
			//}
			//someInternal1 := &PodPeerMatcher{
			//	Namespace: &AllNamespaceMatcher{},
			//	Pod: &AllPodMatcher{},
			//	Port: &AllPortMatcher{},
			//}
			//someInternal2 := &PodPeerMatcher{
			//	Namespace: &AllNamespaceMatcher{},
			//	Pod: &AllPodMatcher{},
			//	Port: &AllPortMatcher{},
			//}
			//
			//Expect(CombinePodPeerMatchers(all, all)).To(Equal(all))
			//Expect(CombinePodPeerMatchers(all, none)).To(Equal(all))
			//Expect(CombinePodPeerMatchers(all, someIps)).To(Equal(all))
			//Expect(CombinePodPeerMatchers(all, someInternal1)).To(Equal(all))
			//Expect(CombinePodPeerMatchers(none, all)).To(Equal(all))
			//Expect(CombinePodPeerMatchers(someIps, all)).To(Equal(all))
			//Expect(CombinePodPeerMatchers(someInternal1, all)).To(Equal(all))
			//
			//Expect(CombinePodPeerMatchers(none, none)).To(Equal(none))
			//Expect(CombinePodPeerMatchers(none, someIps)).To(Equal(someIps))
			//Expect(CombinePodPeerMatchers(none, someInternal1)).To(Equal(someInternal1))
			//Expect(CombinePodPeerMatchers(someIps, none)).To(Equal(someIps))
			//Expect(CombinePodPeerMatchers(someInternal1, none)).To(Equal(someInternal1))
			//
			//bs, err := json.MarshalIndent([]interface{}{someInternal1, someInternal2}, "", "  ")
			//utils.DoOrDie(err)
			//fmt.Printf("%s\n\n", bs)
			//
			//Expect(CombinePodPeerMatchers(someInternal1, someInternal2)).To(Equal(someInternal1))
			//Expect(CombinePodPeerMatchers(someInternal2, someInternal1)).To(Equal(someInternal1))
		})

		It("should combine Internal matchers correctly", func() {
			Expect(nil).ToNot(BeNil())
			//all := &AllInternalMatcher{}
			//none := &NoneInternalMatcher{}
			//np1 := &PodPeerMatcher{
			//	Namespace: &AllNamespaceMatcher{},
			//	Pod:       &LabelSelectorPodMatcher{Selector: *netpol.SelectorAB},
			//	Port:      &AllPortMatcher{},
			//}
			//some1 := &SpecificInternalMatcher{NamespacePods: map[string]*PodPeerMatcher{
			//	np1.PrimaryKey(): np1,
			//}}
			//
			//Expect(CombineInternalMatchers(all, all)).To(Equal(all))
			//Expect(CombineInternalMatchers(all, none)).To(Equal(all))
			//Expect(CombineInternalMatchers(all, some1)).To(Equal(all))
			//Expect(CombineInternalMatchers(none, all)).To(Equal(all))
			//Expect(CombineInternalMatchers(some1, all)).To(Equal(all))
			//
			//Expect(CombineInternalMatchers(none, none)).To(Equal(none))
			//Expect(CombineInternalMatchers(none, some1)).To(Equal(some1))
			//Expect(CombineInternalMatchers(some1, none)).To(Equal(some1))
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
