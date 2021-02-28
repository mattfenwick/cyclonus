package generator

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	. "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func describeDirectionality(isIngress bool) string {
	if isIngress {
		return TagIngress
	} else {
		return TagEgress
	}
}

func describePort(port *intstr.IntOrString) string {
	if port == nil {
		return TagNilPort
	}
	switch port.Type {
	case intstr.Int:
		return TagNumberedPort
	default:
		return TagNamedPort
	}
}

func describeProtocol(protocol *v1.Protocol) string {
	if protocol == nil {
		return TagNilProtocol
	}
	switch *protocol {
	case v1.ProtocolTCP:
		return TagTCPProtocol
	case v1.ProtocolUDP:
		return TagUDPProtocol
	case v1.ProtocolSCTP:
		return TagSCTPProtocol
	default:
		panic(errors.Errorf("invalid protocol %s", *protocol))
	}
}

func (t *TestCaseGeneratorReplacement) ZeroPortProtocolTestCases() []*TestCase {
	var cases []*TestCase
	for _, isIngress := range []bool{false, true} {
		tags := NewStringSet(describeDirectionality(isIngress), TagEmptyPortSlice)
		cases = append(cases, NewSingleStepTestCase("", tags, ProbeAllAvailable,
			CreatePolicy(BuildPolicy(SetPorts(isIngress, emptySliceOfPorts)).NetworkPolicy())))
	}
	return cases
}

func networkPolicyPorts() []NetworkPolicyPort {
	protocols := []*v1.Protocol{
		nil,
		&tcp,
		&udp,
		&sctp,
	}
	ports := []*intstr.IntOrString{
		nil,
		&port80,
		&port81,
	}
	var npps []NetworkPolicyPort
	for _, protocol := range protocols {
		for _, port := range ports {
			npps = append(npps, NetworkPolicyPort{Protocol: protocol, Port: port})
		}
	}
	npps = append(npps,
		NetworkPolicyPort{Protocol: &tcp, Port: &portServe80TCP},
		NetworkPolicyPort{Protocol: &tcp, Port: &portServe81TCP},
		NetworkPolicyPort{Protocol: &udp, Port: &portServe80UDP},
		NetworkPolicyPort{Protocol: &udp, Port: &portServe81UDP},
		NetworkPolicyPort{Protocol: &sctp, Port: &portServe80SCTP},
		NetworkPolicyPort{Protocol: &sctp, Port: &portServe81SCTP})
	return npps
}

func (t *TestCaseGeneratorReplacement) SinglePortProtocolTestCases() []*TestCase {
	var cases []*TestCase
	for _, isIngress := range []bool{false, true} {
		dir := describeDirectionality(isIngress)
		for _, npp := range networkPolicyPorts() {
			tags := NewStringSet(
				TagSinglePortSlice,
				dir,
				describePort(npp.Port),
				describeProtocol(npp.Protocol),
			)
			cases = append(cases, NewSingleStepTestCase("", tags, ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{npp})).NetworkPolicy())))
		}

		// pathological cases
		cases = append(cases,
			NewSingleStepTestCase("open a named port that doesn't match its protocol",
				NewStringSet(TagSinglePortSlice, TagPathological, dir, describePort(&portServe81UDP), describeProtocol(&tcp)),
				ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &tcp, Port: &portServe81UDP}})).NetworkPolicy())),
			NewSingleStepTestCase("open a named port that isn't served",
				NewStringSet(TagSinglePortSlice, TagPathological, dir, describePort(&portServe7981UDP), describeProtocol(&tcp)),
				ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &tcp, Port: &portServe7981UDP}})).NetworkPolicy())),
			NewSingleStepTestCase("open a numbered port that isn't served",
				NewStringSet(TagSinglePortSlice, TagPathological, dir, describePort(&port7981), describeProtocol(&tcp)),
				ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &tcp, Port: &port7981}})).NetworkPolicy())))
	}
	return cases
}

func (t *TestCaseGeneratorReplacement) TwoPortProtocolTestCases() []*TestCase {
	var cases []*TestCase
	for _, isIngress := range []bool{false, true} {
		for i, ports1 := range networkPolicyPorts() {
			for j, ports2 := range networkPolicyPorts() {
				if i < j {
					tags := NewStringSet(
						TagTwoPlusPortSlice,
						describeDirectionality(isIngress),
						describePort(ports1.Port),
						describeProtocol(ports1.Protocol),
						describePort(ports2.Port),
						describeProtocol(ports2.Protocol))
					cases = append(cases, NewSingleStepTestCase("", tags, ProbeAllAvailable,
						CreatePolicy(BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{ports1, ports2})).NetworkPolicy())))
				}
			}
		}
	}
	return cases
}

func (t *TestCaseGeneratorReplacement) PortProtocolTestCases() []*TestCase {
	var cases []*TestCase
	cases = append(cases, t.ZeroPortProtocolTestCases()...)
	cases = append(cases, t.SinglePortProtocolTestCases()...)
	cases = append(cases, t.TwoPortProtocolTestCases()...)
	return cases
}
