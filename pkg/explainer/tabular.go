package explainer

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"strings"
)

type SliceBuilder struct {
	Prefix   []string
	Elements [][]string
}

func (s *SliceBuilder) Append(items ...string) {
	s.Elements = append(s.Elements, append(s.Prefix, items...))
}

func TableExplainer(policies *matcher.Policy) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Type", "Target", "Source rules", "Peer", "Port/Protocol"})

	builder := &SliceBuilder{}
	ingresses, egresses := policies.SortedTargets()
	TargetsTableLines(builder, ingresses, true)
	builder.Elements = append(builder.Elements, []string{"", "", "", "", ""})
	TargetsTableLines(builder, egresses, false)

	table.AppendBulk(builder.Elements)

	table.Render()
	return tableString.String()
}

func TargetsTableLines(builder *SliceBuilder, targets []*matcher.Target, isIngress bool) {
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
		target := fmt.Sprintf("namespace: %s\n%s", ingress.Namespace, kube.LabelSelectorTableLines(ingress.PodSelector))
		rules := strings.Join(sourceRules, "\n")
		builder.Prefix = []string{ruleType, target, rules}

		switch a := ingress.Peer.(type) {
		case *matcher.AllPeerMatcher:
			builder.Append("all pods, all ips", "all ports, all protocols")
		case *matcher.NonePeerMatcher:
			builder.Append("no pods, no ips", "no ports, no protocols")
		case *matcher.SpecificPeerMatcher:
			switch ip := a.IP.(type) {
			case *matcher.AllIPMatcher:
				builder.Append("all ips", "all ports, all protocols")
			case *matcher.NoneIPMatcher:
				builder.Append("no ips", "no ports, no protocols")
			case *matcher.SpecificIPMatcher:
				SpecificIPMatcherTableLines(builder, ip)
			default:
				panic(errors.Errorf("invalid IPMatcher type %T", ip))
			}
			switch internal := a.Internal.(type) {
			case *matcher.AllInternalMatcher:
				builder.Append("all pods", "all ports, all protocols")
			case *matcher.NoneInternalMatcher:
				builder.Append("no pods", "no ports, no protocols")
			case *matcher.SpecificInternalMatcher:
				SpecificInternalMatcherTableLines(builder, internal)
			default:
				panic(errors.Errorf("invalid InternalMatcher type %T", internal))
			}
		default:
			panic(errors.Errorf("invalid PeerMatcher type %T", a))
		}
	}
}

func SpecificIPMatcherTableLines(builder *SliceBuilder, ip *matcher.SpecificIPMatcher) {
	builder.Append("ports for all IPs", strings.Join(PortMatcherTableLines(ip.PortsForAllIPs), "\n"))
	for _, block := range ip.SortedIPBlocks() {
		peer := block.IPBlock.CIDR + "\n" + fmt.Sprintf("except %+v", block.IPBlock.Except)
		pps := PortMatcherTableLines(block.Port)
		builder.Append(peer, strings.Join(pps, "\n"))
	}
}

func SpecificInternalMatcherTableLines(builder *SliceBuilder, internal *matcher.SpecificInternalMatcher) {
	for _, nsPodMatcher := range internal.NamespacePods {
		var namespaces string
		switch ns := nsPodMatcher.Namespace.(type) {
		case *matcher.AllNamespaceMatcher:
			namespaces = "all"
		case *matcher.LabelSelectorNamespaceMatcher:
			namespaces = kube.LabelSelectorTableLines(ns.Selector)
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
			pods = kube.LabelSelectorTableLines(p.Selector)
		default:
			panic(errors.Errorf("invalid PodMatcher type %T", p))
		}
		builder.Append("namespace: "+namespaces+"\n"+"pods: "+pods, strings.Join(PortMatcherTableLines(nsPodMatcher.Port), "\n"))
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
