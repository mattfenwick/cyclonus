package explainer

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func TableExplainer(policies *matcher.Policy) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Type", "Target", "Source rules", "Peer", "Port/Protocol"})

	ingresses, egresses := policies.SortedTargets()
	TargetsTableLines(table, ingresses, true)
	TargetsTableLines(table, egresses, false)

	table.Render()
	return tableString.String()
}

func row(policyType, target, sourceRules, peer, portProtocol string) []string {
	return []string{policyType, target, sourceRules, peer, portProtocol}
}

func TargetsTableLines(table *tablewriter.Table, targets []*matcher.Target, isIngress bool) {
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
		target := fmt.Sprintf("namespace: %s\n%s", ingress.Namespace, LabelSelectorTableLines(ingress.PodSelector))
		rules := strings.Join(sourceRules, "\n")

		switch a := ingress.Peer.(type) {
		case *matcher.AllPeerMatcher:
			table.Append(row(ruleType, target, rules, "all pods, all ips", "all ports, all protocols"))
		case *matcher.NonePeerMatcher:
			table.Append(row(ruleType, target, rules, "no pods, no ips", "no ports, no protocols"))
		case *matcher.SpecificPeerMatcher:
			switch ip := a.IP.(type) {
			case *matcher.AllIPMatcher:
				table.Append(row(ruleType, target, rules, "all ips", "all ports, all protocols"))
			case *matcher.NoneIPMatcher:
				table.Append(row(ruleType, target, rules, "no ips", "no ports, no protocols"))
			case *matcher.SpecificIPMatcher:
				table.Append(row(ruleType, target, rules, "ports for all IPs", strings.Join(PortMatcherTableLines(ip.PortsForAllIPs), "\n")))
				for _, block := range ip.SortedIPBlocks() {
					pps := PortMatcherTableLines(block.Port)
					table.Append(row(
						ruleType,
						target,
						rules,
						strings.Join(append([]string{block.IPBlock.CIDR}, fmt.Sprintf("except %+v", block.IPBlock.Except)), "\n"),
						strings.Join(pps, "\n")))
				}
			default:
				panic(errors.Errorf("invalid IPMatcher type %T", ip))
			}
			switch internal := a.Internal.(type) {
			case *matcher.AllInternalMatcher:
				table.Append(row(ruleType, target, rules, "all pods", "all ports, all protocols"))
			case *matcher.NoneInternalMatcher:
				table.Append(row(ruleType, target, rules, "no pods", "no ports, no protocols"))
			case *matcher.SpecificInternalMatcher:
				for _, nsPodMatcher := range internal.NamespacePods {
					var namespaces string
					switch ns := nsPodMatcher.Namespace.(type) {
					case *matcher.AllNamespaceMatcher:
						namespaces = "all"
					case *matcher.LabelSelectorNamespaceMatcher:
						namespaces = LabelSelectorTableLines(ns.Selector)
					case *matcher.ExactNamespaceMatcher:
						namespaces = ns.Namespace
					default:
						panic(errors.Errorf("invalid NamespaceMatcher type %T", ns))
					}
					var pods string
					switch p := nsPodMatcher.Pod.(type) {
					case *matcher.AllPodMatcher:
						pods = "all"
					case *matcher.LabelSelectorPodMatcher:
						pods = LabelSelectorTableLines(p.Selector)
					default:
						panic(errors.Errorf("invalid PodMatcher type %T", p))
					}
					table.Append(row(ruleType,
						target,
						rules,
						"namespace: "+namespaces+"\n"+"pods: "+pods,
						strings.Join(PortMatcherTableLines(nsPodMatcher.Port), "\n")))
				}
			default:
				panic(errors.Errorf("invalid InternalMatcher type %T", internal))
			}
		default:
			panic(errors.Errorf("invalid PeerMatcher type %T", a))
		}
	}
}

func PortMatcherTableLines(pm matcher.PortMatcher) []string {
	switch port := pm.(type) {
	case *matcher.AllPortMatcher:
		return []string{"all ports, all protocols"}
	case *matcher.NonePortMatcher:
		return []string{"no ports, no protocols"}
	case *matcher.SpecificPortMatcher:
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
	if kube.IsLabelSelectorEmpty(selector) {
		return "all pods"
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
