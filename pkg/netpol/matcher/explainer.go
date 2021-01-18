package matcher

import (
	"fmt"
	"github.com/pkg/errors"
	networkingv1 "k8s.io/api/networking/v1"
	"strings"
)

func Explain(policies *Policy) string {
	var lines []string
	// 1. ingress
	for _, t := range policies.Ingress {
		lines = append(lines, ExplainTarget(t, true)...)
	}

	// 2. egress
	for _, t := range policies.Egress {
		lines = append(lines, ExplainTarget(t, false)...)
	}

	return strings.Join(lines, "\n")
}

func ExplainTarget(target *Target, isIngress bool) []string {
	indent := "  "
	var targetType string
	if isIngress {
		targetType = "ingress"
	} else {
		targetType = "egress"
	}
	var lines []string
	lines = append(lines, target.GetPrimaryKey())
	if len(target.SourceRules) != 0 {
		lines = append(lines, indent+"source rules:")
		lines = append(lines, ExplainSourceRules(target.SourceRules, indent+"  ")...)
	}
	switch a := target.Peer.(type) {
	case *NonePeerMatcher:
		lines = append(lines, fmt.Sprintf(indent+"all %s blocked", targetType))
	case *AllPeerMatcher:
		lines = append(lines, fmt.Sprintf(indent+"all %s allowed", targetType))
	case *SpecificPeerMatcher:
		lines = append(lines, fmt.Sprintf(indent+"%s:", targetType))
		lines = append(lines, ExplainSpecificPeerMatcher(a, indent+"  ")...)
	default:
		panic(errors.Errorf("invalid PeerMatcher type %T", target.Peer))
	}

	lines = append(lines, "")
	return lines
}

func ExplainSourceRules(sourceRules []*networkingv1.NetworkPolicy, indent string) []string {
	var lines []string
	for _, sr := range sourceRules {
		lines = append(lines, fmt.Sprintf(indent+"%s/%s", sr.Namespace, sr.Name))
	}
	return lines
}

func ExplainSpecificPeerMatcher(tp *SpecificPeerMatcher, indent string) []string {
	lines := ExplainIPMatcher(tp.IP, indent)
	return append(lines, ExplainInternalMatcher(tp.Internal, indent)...)
}

func ExplainIPMatcher(ip IPMatcher, indent string) []string {
	switch a := ip.(type) {
	case *AllIPMatcher:
		return []string{indent + "all ips"}
	case *NoneIPMatcher:
		return []string{indent + "no ips"}
	case *SpecificIPMatcher:
		lines := []string{indent + "Ports for all IPs"}
		lines = append(lines, ExplainPortMatcher(a.PortsForAllIPs, indent+"  ")...)
		lines = append(lines, indent+"IPBlock(s):")
		// TODO ensure consistent order
		for _, ip := range a.IPBlocks {
			lines = append(lines, ExplainIPBlockMatcher(ip, indent+"  ")...)
		}
		return lines
	default:
		panic(errors.Errorf("invalid IPMatcher type %T", ip))
	}
}

func ExplainIPBlockMatcher(ip *IPBlockMatcher, indent string) []string {
	var lines []string
	block := fmt.Sprintf("IPBlock: cidr %s, except %+v", ip.IPBlock.CIDR, ip.IPBlock.Except)
	lines = append(lines, indent+block)
	for _, port := range ExplainPortMatcher(ip.Port, indent+"  ") {
		lines = append(lines, port)
	}
	return lines
}

func ExplainPortMatcher(pm PortMatcher, indent string) []string {
	lines := []string{indent + "Port(s):"}
	switch m := pm.(type) {
	case *NonePortMatcher:
		return append(lines, indent+"no ports")
	case *AllPortMatcher:
		return append(lines, ExplainAllPortMatcher(indent+"  ")...)
	case *SpecificPortMatcher:
		return append(lines, ExplainSpecificPortMatcher(m, indent+"  ")...)
	default:
		panic(errors.Errorf("invalid Port type %T", pm))
	}
}

func ExplainAllPortMatcher(indent string) []string {
	return []string{indent + "all ports all protocols"}
}

func ExplainSpecificPortMatcher(spm *SpecificPortMatcher, indent string) []string {
	var lines []string
	for _, port := range spm.Ports {
		if port.Port != nil {
			lines = append(lines, indent+fmt.Sprintf("port %s on protocol %s", port.Port.String(), port.Protocol))
		} else {
			lines = append(lines, indent+fmt.Sprintf("all ports on protocol %s", port.Protocol))
		}
	}
	return lines
}

func ExplainInternalMatcher(i InternalMatcher, indent string) []string {
	lines := []string{indent + "Internal:"}
	switch l := i.(type) {
	case *NoneInternalMatcher:
		lines = append(lines, indent+"all pods blocked")
	case *AllInternalMatcher:
		lines = append(lines, indent+"all pods in all namespaces")
	case *SpecificInternalMatcher:
		// TODO ensure consistent order
		for _, peer := range l.NamespacePods {
			lines = append(lines, ExplainNamespacePod(peer, indent+"  ")...)
		}
	}
	return lines
}

func ExplainNamespacePod(peer *NamespacePodMatcher, indent string) []string {
	lines := []string{indent + "Namespace/Pod:"}
	lines = append(lines, ExplainNamespaceMatcher(peer.Namespace, indent+"  "), ExplainPodMatcher(peer.Pod, indent+"  "))
	for _, port := range ExplainPortMatcher(peer.Port, indent+"  ") {
		lines = append(lines, port)
	}
	return lines
}

func ExplainPodMatcher(pm PodMatcher, indent string) string {
	switch m := pm.(type) {
	case *AllPodMatcher:
		return indent + "all pods"
	case *LabelSelectorPodMatcher:
		return indent + "pods matching " + SerializeLabelSelector(m.Selector)
	default:
		panic(errors.Errorf("invalid PodMatcher type %T", pm))
	}
}

func ExplainNamespaceMatcher(pm NamespaceMatcher, indent string) string {
	switch m := pm.(type) {
	case *AllNamespaceMatcher:
		return indent + "all namespaces"
	case *ExactNamespaceMatcher:
		return indent + "namespace " + m.Namespace
	case *LabelSelectorNamespaceMatcher:
		return indent + "namespaces matching " + SerializeLabelSelector(m.Selector)
	default:
		panic(errors.Errorf("invalid NamespaceMatcher type %T", pm))
	}
}
