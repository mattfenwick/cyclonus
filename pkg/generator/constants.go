package generator

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	sctp = v1.ProtocolSCTP
	tcp  = v1.ProtocolTCP
	udp  = v1.ProtocolUDP

	port53 = intstr.FromInt(53)

	port79   = intstr.FromInt(79)
	port80   = intstr.FromInt(80)
	port81   = intstr.FromInt(81)
	port82   = intstr.FromInt(82)
	port7981 = intstr.FromInt(7981)
)

var (
	portServe79TCP = intstr.FromString("serve-79-tcp")
	portServe80TCP = intstr.FromString("serve-80-tcp")
	portServe81TCP = intstr.FromString("serve-81-tcp")

	portServe80UDP   = intstr.FromString("serve-80-udp")
	portServe81UDP   = intstr.FromString("serve-81-udp")
	portServe7981UDP = intstr.FromString("serve-7981-udp")

	portServe80SCTP = intstr.FromString("serve-80-sctp")
	portServe81SCTP = intstr.FromString("serve-81-sctp")
)

var (
	ProbeAllAvailable = &ProbeConfig{AllAvailable: true}

	probePort80TCP = &ProbeConfig{PortProtocol: &PortProtocol{Port: port80, Protocol: tcp}}
	probePort81TCP = &ProbeConfig{PortProtocol: &PortProtocol{Port: port81, Protocol: tcp}}

	probePortServe80TCP = &ProbeConfig{PortProtocol: &PortProtocol{Port: portServe80TCP, Protocol: tcp}}
	probePortServe81TCP = &ProbeConfig{PortProtocol: &PortProtocol{Port: portServe81TCP, Protocol: tcp}}
)

var (
	DenyAllRules  = []*Rule{}
	AllowAllRules = []*Rule{{}}
)
