package generator

import (
	"fmt"
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
		return TagAnyPort
	}
	switch port.Type {
	case intstr.Int:
		return TagNumberedPort
	default:
		return TagNamedPort
	}
}

func describeProtocol(protocol *v1.Protocol) *string {
	if protocol == nil {
		return nil
	}
	var tag string
	switch *protocol {
	case v1.ProtocolTCP:
		tag = TagTCPProtocol
	case v1.ProtocolUDP:
		tag = TagUDPProtocol
	case v1.ProtocolSCTP:
		tag = TagSCTPProtocol
	default:
		panic(errors.Errorf("invalid protocol %s", *protocol))
	}
	return &tag
}

func (t *TestCaseGenerator) ZeroPortProtocolTestCases() []*TestCase {
	var cases []*TestCase
	for _, isIngress := range []bool{false, true} {
		dir := describeDirectionality(isIngress)
		tags := NewStringSet(dir, TagAnyPortProtocol)
		cases = append(cases, NewSingleStepTestCase(fmt.Sprintf("%s: empty port/protocol", dir), tags, ProbeAllAvailable,
			CreatePolicy(BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{})).NetworkPolicy())))
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

func (t *TestCaseGenerator) SinglePortProtocolTestCases() []*TestCase {
	var cases []*TestCase
	for _, isIngress := range []bool{false, true} {
		dir := describeDirectionality(isIngress)
		for _, npp := range networkPolicyPorts() {
			tags := NewStringSet(
				dir,
				describePort(npp.Port),
			)
			if tag := describeProtocol(npp.Protocol); tag != nil {
				tags.Add(*tag)
			}
			cases = append(cases, NewSingleStepTestCase("", tags, ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{npp})).NetworkPolicy())))
		}

		// pathological cases
		cases = append(cases,
			NewSingleStepTestCase("open a named port that doesn't match its protocol",
				NewStringSet(TagPathological, dir, describePort(&portServe81UDP), TagTCPProtocol),
				ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &tcp, Port: &portServe81UDP}})).NetworkPolicy())),
			NewSingleStepTestCase("open a named port that isn't served",
				NewStringSet(TagPathological, dir, describePort(&portServe7981UDP), TagTCPProtocol),
				ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &tcp, Port: &portServe7981UDP}})).NetworkPolicy())),
			NewSingleStepTestCase("open a numbered port that isn't served",
				NewStringSet(TagPathological, dir, describePort(&port7981), TagTCPProtocol),
				ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{{Protocol: &tcp, Port: &port7981}})).NetworkPolicy())))
	}
	return cases
}

func (t *TestCaseGenerator) TwoPortProtocolTestCases() []*TestCase {
	// TODO better way to handle this?  this generates too many, and most don't seem helpful ... but
	//var cases []*TestCase
	//for _, isIngress := range []bool{false, true} {
	//	for i, ports1 := range networkPolicyPorts() {
	//		for j, ports2 := range networkPolicyPorts() {
	//			if i < j {
	//				tags := NewStringSet(
	//					TagTwoPlusPortSlice,
	//					describeDirectionality(isIngress),
	//					describePort(ports1.Port),
	//					describeProtocol(ports1.Protocol),
	//					describePort(ports2.Port),
	//					describeProtocol(ports2.Protocol))
	//				cases = append(cases, NewSingleStepTestCase("", tags, ProbeAllAvailable,
	//					CreatePolicy(BuildPolicy(SetPorts(isIngress, []NetworkPolicyPort{ports1, ports2})).NetworkPolicy())))
	//			}
	//		}
	//	}
	//}
	//return cases
	// all/all, implicit/explicit protocol, all ports, numbered port, named port
	nppPairs := [][]NetworkPolicyPort{
		{{}, {Port: &port80}},
		{{}, {Port: &portServe80TCP}},
		{{}, {Protocol: &udp}},
		{{Port: &port80}, {Port: &port81}},
		{{Port: &port80}, {Port: &portServe81TCP}},
		{{Port: &port80}, {Protocol: &udp, Port: &portServe81UDP}},
		{{Protocol: &udp, Port: &port80}, {Protocol: &udp, Port: &portServe81UDP}},
	}
	var cases []*TestCase
	for _, isIngress := range []bool{false, true} {
		dir := describeDirectionality(isIngress)
		for _, nppSlice := range nppPairs {
			tags := NewStringSet(TagMultiPortProtocol, dir)
			for _, pp := range nppSlice {
				if tag := describeProtocol(pp.Protocol); tag != nil {
					tags.Add(*tag)
				}
				tags.Add(describePort(pp.Port))
			}
			cases = append(cases, NewSingleStepTestCase("", tags, ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPorts(isIngress, nppSlice)).NetworkPolicy())))
		}
	}
	return cases
}

func (t *TestCaseGenerator) PortProtocolTestCases() []*TestCase {
	var cases []*TestCase
	cases = append(cases, t.ZeroPortProtocolTestCases()...)
	cases = append(cases, t.SinglePortProtocolTestCases()...)
	cases = append(cases, t.TwoPortProtocolTestCases()...)
	return cases
}
