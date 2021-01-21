package netpol

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	SCTP        = v1.ProtocolSCTP
	TCP         = v1.ProtocolTCP
	UDP         = v1.ProtocolUDP
	Port53      = intstr.FromInt(53)
	Port80      = intstr.FromInt(80)
	Port443     = intstr.FromInt(443)
	Port988     = intstr.FromInt(988)
	Port9001Ref = intstr.FromInt(9001)
	PortHello   = intstr.FromString("hello")
)
