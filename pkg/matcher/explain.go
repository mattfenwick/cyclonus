package matcher

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
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

func (p *Policy) ExplainTable() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Type", "Target", "Source rules", "Peer", "Port/Protocol"})

	builder := &SliceBuilder{}
	ingresses, egresses := p.SortedTargets()
	builder.TargetsTableLines(ingresses, true)
	builder.Elements = append(builder.Elements, []string{"", "", "", "", ""})
	builder.TargetsTableLines(egresses, false)

	table.AppendBulk(builder.Elements)

	table.Render()
	return tableString.String()
}

func (s *SliceBuilder) TargetsTableLines(targets []*Target, isIngress bool) {
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
		s.Prefix = []string{ruleType, target, rules}

		switch a := ingress.Peer.(type) {
		case *AllPeerMatcher:
			s.Append("all pods, all ips", "all ports, all protocols")
		case *NonePeerMatcher:
			s.Append("no pods, no ips", "no ports, no protocols")
		case *SpecificPeerMatcher:
			switch ip := a.IP.(type) {
			case *AllIPMatcher:
				s.Append("all ips", "all ports, all protocols")
			case *NoneIPMatcher:
				s.Append("no ips", "no ports, no protocols")
			case *SpecificIPMatcher:
				s.SpecificIPMatcherTableLines(ip)
			default:
				panic(errors.Errorf("invalid IPMatcher type %T", ip))
			}
			switch internal := a.Internal.(type) {
			case *AllInternalMatcher:
				s.Append("all pods", "all ports, all protocols")
			case *NoneInternalMatcher:
				s.Append("no pods", "no ports, no protocols")
			case *SpecificInternalMatcher:
				s.SpecificInternalMatcherTableLines(internal)
			default:
				panic(errors.Errorf("invalid InternalMatcher type %T", internal))
			}
		default:
			panic(errors.Errorf("invalid PeerMatcher type %T", a))
		}
	}
}

func (s *SliceBuilder) SpecificIPMatcherTableLines(ip *SpecificIPMatcher) {
	s.Append("ports for all IPs", strings.Join(PortMatcherTableLines(ip.PortsForAllIPs), "\n"))
	for _, block := range ip.SortedIPBlocks() {
		peer := block.IPBlock.CIDR + "\n" + fmt.Sprintf("except %+v", block.IPBlock.Except)
		pps := PortMatcherTableLines(block.Port)
		s.Append(peer, strings.Join(pps, "\n"))
	}
}

func (s *SliceBuilder) SpecificInternalMatcherTableLines(internal *SpecificInternalMatcher) {
	for _, nsPodMatcher := range internal.NamespacePods {
		var namespaces string
		switch ns := nsPodMatcher.Namespace.(type) {
		case *AllNamespaceMatcher:
			namespaces = "all"
		case *LabelSelectorNamespaceMatcher:
			namespaces = kube.LabelSelectorTableLines(ns.Selector)
		case *ExactNamespaceMatcher:
			namespaces = ns.Namespace
		default:
			panic(errors.Errorf("invalid NamespaceMatcher type %T", ns))
		}
		var pods string
		switch p := nsPodMatcher.Pod.(type) {
		case *AllPodMatcher:
			pods = "all"
		case *LabelSelectorPodMatcher:
			pods = kube.LabelSelectorTableLines(p.Selector)
		default:
			panic(errors.Errorf("invalid PodMatcher type %T", p))
		}
		s.Append("namespace: "+namespaces+"\n"+"pods: "+pods, strings.Join(PortMatcherTableLines(nsPodMatcher.Port), "\n"))
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
