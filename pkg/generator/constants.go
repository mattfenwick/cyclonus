package generator

import (
	v1 "k8s.io/api/core/v1"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	ProbeAllAvailable = &ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName}

	probePort80TCP = NewProbeConfig(port80, tcp, ProbeModeServiceName)
	probePort81TCP = NewProbeConfig(port81, tcp, ProbeModeServiceName)

	probePortServe80TCP = NewProbeConfig(portServe80TCP, tcp, ProbeModeServiceName)
	probePortServe81TCP = NewProbeConfig(portServe81TCP, tcp, ProbeModeServiceName)
)

var (
	DenyAllRules  = []*Rule{}
	AllowAllRules = []*Rule{{}}
)

var (
	AllowDNSRule = &Rule{
		Ports: []NetworkPolicyPort{
			{
				Protocol: &udp,
				Port:     &port53,
			},
			{
				Protocol: &tcp,
				Port:     &port53,
			},
		},
	}

	AllowDNSPeers = &NetpolPeers{
		Rules: []*Rule{AllowDNSRule},
	}
)

func AllowDNSPolicy(source *NetpolTarget) *Netpol {
	return &Netpol{
		Name:   "allow-dns",
		Target: source,
		Egress: AllowDNSPeers,
	}
}

var (
	emptySelector                 = &metav1.LabelSelector{}
	podAMatchLabelsSelector       = &metav1.LabelSelector{MatchLabels: map[string]string{"pod": "a"}}
	podCMatchLabelsSelector       = &metav1.LabelSelector{MatchLabels: map[string]string{"pod": "c"}}
	podABMatchExpressionsSelector = &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "pod",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"a", "b"},
			},
		},
	}
	podBCMatchExpressionsSelector = &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "pod",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"b", "c"},
			},
		},
	}

	nsXMatchLabelsSelector       = &metav1.LabelSelector{MatchLabels: map[string]string{"ns": "x"}}
	nsXYMatchExpressionsSelector = &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "ns",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"x", "y"},
			},
		},
	}
	nsYZMatchExpressionsSelector = &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "ns",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"y", "z"},
			},
		},
	}
)
