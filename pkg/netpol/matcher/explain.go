package matcher

import (
	"fmt"
	"github.com/pkg/errors"
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
	var targetType string
	if isIngress {
		targetType = "ingress"
	} else {
		targetType = "egress"
	}
	var lines []string
	lines = append(lines, target.GetPrimaryKey())
	if len(target.SourceRules) != 0 {
		lines = append(lines, "  source rules:")
		for _, sr := range target.SourceRules {
			lines = append(lines, fmt.Sprintf("    %s/%s", sr.Namespace, sr.Name))
		}
	}
	switch a := target.Peer.(type) {
	case *NonePeerMatcher:
		lines = append(lines, fmt.Sprintf("  all %s blocked", targetType))
	case *AllPeerMatcher:
		lines = append(lines, fmt.Sprintf("  all %s allowed", targetType))
	case *SpecificPeerMatcher:
		lines = append(lines, fmt.Sprintf("  %s:", targetType))
		lines = append(lines, ExplainEdgePeerPortMatcher(a)...)
	default:
		panic(errors.Errorf("invalid PeerMatcher type %T", target.Peer))
	}

	lines = append(lines, "")
	return lines
}

func ExplainEdgePeerPortMatcher(tp *SpecificPeerMatcher) []string {
	var lines []string
	for _, ip := range tp.IP {
		block := fmt.Sprintf("IPBlock: cidr %s, except %+v", ip.IPBlock.CIDR, ip.IPBlock.Except)
		lines = append(lines, fmt.Sprintf("  - %s", block))
		for _, port := range ExplainPortMatcher(ip.Port) {
			lines = append(lines, "    "+port)
		}
	}
	return append(lines, ExplainInternal(tp.Internal)...)
}

func ExplainPortMatcher(pm PortMatcher) []string {
	switch m := pm.(type) {
	case *AllPortMatcher:
		return []string{"all ports all protocols"}
	case *SpecificPortsMatcher:
		var lines []string
		for _, port := range m.Ports {
			lines = append(lines)
			if port.Port != nil {
				lines = append(lines, fmt.Sprintf("port %s on protocol %s", port.Port.String(), port.Protocol))
			} else {
				lines = append(lines, fmt.Sprintf("all ports on protocol %s", port.Protocol))
			}
		}
		return lines
	default:
		panic(errors.Errorf("invalid Port type %T", pm))
	}
}

func ExplainPodMatcher(pm PodMatcher) string {
	switch m := pm.(type) {
	case *AllPodMatcher:
		return "all pods"
	case *LabelSelectorPodMatcher:
		return "pods matching " + SerializeLabelSelector(m.Selector)
	default:
		panic(errors.Errorf("invalid PodMatcher type %T", pm))
	}
}

func ExplainNamespaceMatcher(pm NamespaceMatcher) string {
	switch m := pm.(type) {
	case *AllNamespaceMatcher:
		return "all namespaces"
	case *ExactNamespaceMatcher:
		return "namespace " + m.Namespace
	case *LabelSelectorNamespaceMatcher:
		return "namespaces matching " + SerializeLabelSelector(m.Selector)
	default:
		panic(errors.Errorf("invalid NamespaceMatcher type %T", pm))
	}
}

func ExplainInternal(i InternalMatcher) []string {
	var lines []string
	switch l := i.(type) {
	case *NoneInternalMatcher:
		lines = append(lines, "  all pods blocked")
	case *AllInternalMatcher:
		lines = append(lines, "    all pods in all namespaces")
	case *SpecificInternalMatcher:
		for _, peer := range l.Pods {
			lines = append(lines, fmt.Sprintf("    %s; %s", ExplainNamespaceMatcher(peer.Namespace), ExplainPodMatcher(peer.Pod)))
			for _, port := range ExplainPortMatcher(peer.Port) {
				lines = append(lines, "      "+port)
			}
		}
	}
	return lines
}
