package matcher

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strings"
)

func TableExplainer(policies *Policy) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetHeader([]string{"Type", "Target namespace", "Target pod selector", "Peer", "Port/Protocol", "Source rules"})

	ingresses, egresses := policies.SortedTargets()
	TargetsTableLines(table, ingresses, true)
	TargetsTableLines(table, egresses, false)
	return table
}

func TargetsTableLines(table *tablewriter.Table, targets []*Target, isIngress bool) {
	var ruleType string
	if isIngress {
		ruleType = "Ingress"
	} else {
		ruleType = "Egress"
	}
	for _, ingress := range targets {
		var sourceRules []string
		for _, sr := range ingress.SourceRules {
			sourceRules = append(sourceRules, fmt.Sprintf("%s/%s", sr.Namespace, sr.Name))
		}
		table.Append([]string{
			ruleType,
			ingress.Namespace,
			LabelSelectorTableLines(ingress.PodSelector),
			"", // peer,
			"", // port/protocol,
			strings.Join(sourceRules, "\n"),
		})
		switch a := ingress.Peer.(type) {
		case *AllPeerMatcher:
			table.Append([]string{"", "", "", "all pods, all ips", "all ports, all protocols", ""})
		case *NonePeerMatcher:
			table.Append([]string{"", "", "", "no pods, no ips", "no ports, no protocols", ""})
		case *SpecificPeerMatcher:
			switch ip := a.IP.(type) {
			case *AllIPMatcher:
				table.Append([]string{"", "", "", "all ips", "all ports, all protocols", ""})
			case *NoneIPMatcher:
				table.Append([]string{"", "", "", "no ips", "no ports, no protocols", ""})
			case *SpecificIPMatcher:
				table.Append([]string{"", "", "", "ports for all IPs", strings.Join(PortMatcherTableLines(ip.PortsForAllIPs), "\n"), ""})
				for _, block := range ip.SortedIPBlocks() {
					pps := PortMatcherTableLines(block.Port)
					table.Append([]string{
						"",
						"",
						"",
						strings.Join(append([]string{block.IPBlock.CIDR}, block.IPBlock.Except...), "\n"),
						strings.Join(pps, "\n"),
						"",
					})
				}
			default:
				panic(errors.Errorf("invalid IPMatcher type %T", ip))
			}
			switch internal := a.Internal.(type) {
			case *AllInternalMatcher:
			case *NoneInternalMatcher:
			case *SpecificInternalMatcher:
				for _, matcher := range internal.NamespacePods {
					var namespaces string
					switch ns := matcher.Namespace.(type) {
					case *AllNamespaceMatcher:
						namespaces = "all"
					case *LabelSelectorNamespaceMatcher:
						namespaces = LabelSelectorTableLines(ns.Selector)
					case *ExactNamespaceMatcher:
						namespaces = ns.Namespace
					default:
						panic(errors.Errorf("invalid NamespaceMatcher type %T", ns))
					}
					var pods string
					switch p := matcher.Pod.(type) {
					case *AllPodMatcher:
						pods = "all"
					case *LabelSelectorPodMatcher:
						pods = LabelSelectorTableLines(p.Selector)
					default:
						panic(errors.Errorf("invalid PodMatcher type %T", p))
					}
					table.Append([]string{
						"",
						"",
						"",
						"namespace: " + namespaces + "\n" + "pods: " + pods,
						strings.Join(PortMatcherTableLines(matcher.Port), "\n"),
						"",
					})
				}
			default:
				panic(errors.Errorf("invalid InternalMatcher type %T", internal))
			}
		default:
			panic(errors.Errorf("invalid PeerMatcher type %T", a))
		}
	}
}

func PortMatcherTableLines(pm PortMatcher) []string {
	switch port := pm.(type) {
	case *AllPortMatcher:
		return []string{"all ports, all protocols"}
	case *NonePortMatcher:
		return []string{"no ports, no protocols"}
	case *SpecificPortMatcher:
		var pps []string
		for _, portProtocol := range port.Ports {
			if portProtocol.Port == nil {
				pps = append(pps, "all ports on protocol "+string(portProtocol.Protocol))
			} else {
				pps = append(pps, "port "+portProtocol.Port.String()+" on protocol "+string(portProtocol.Protocol))
			}
		}
		return pps
	default:
		panic(errors.Errorf("invalid PortMatcher type %T", port))
	}
}

func LabelSelectorTableLines(selector metav1.LabelSelector) string {
	if isLabelSelectorEmpty(selector) {
		return "all"
	}
	var lines []string
	if len(selector.MatchLabels) > 0 {
		lines = append(lines, "Match labels:")
		for key, val := range selector.MatchLabels {
			lines = append(lines, fmt.Sprintf("  %s: %s", key, val))
		}
	}
	if len(selector.MatchExpressions) > 0 {
		lines = append(lines, "Match expressions:")
		for _, exp := range selector.MatchExpressions {
			lines = append(lines, fmt.Sprintf("  %s %s %+v", exp.Key, exp.Operator, exp.Values))
		}
	}
	return strings.Join(lines, "\n")
}

func Explain(policies *Policy) string {
	var lines []string
	ingress, egress := policies.SortedTargets()
	// 1. ingress
	for _, t := range ingress {
		lines = append(lines, ExplainTarget(t, true)...)
	}

	// 2. egress
	for _, t := range egress {
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
		for _, ip := range a.SortedIPBlocks() {
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
		for _, peer := range l.SortedNamespacePods() {
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
