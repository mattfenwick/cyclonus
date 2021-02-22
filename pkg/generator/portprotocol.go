package generator

import (
	v1 "k8s.io/api/core/v1"
	. "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type PortProtocol struct {
	Protocol v1.Protocol
	Port     intstr.IntOrString
}

func SinglePortProtocolTestCases() []NetworkPolicyPort {
	return []NetworkPolicyPort{
		{Protocol: nil, Port: nil},
		{Protocol: &tcp, Port: nil},
		{Protocol: nil, Port: &port80},
		{Protocol: &tcp, Port: &port80},
		{Protocol: nil, Port: &portServe81TCP},
		{Protocol: &tcp, Port: &portServe81TCP},
	}
}
