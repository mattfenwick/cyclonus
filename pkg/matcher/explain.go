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
	for _, target := range targets {
		var sourceRules []string
		for _, sr := range target.SourceRules {
			sourceRules = append(sourceRules, fmt.Sprintf("%s/%s", sr.Namespace, sr.Name))
		}
		targetString := fmt.Sprintf("namespace: %s\n%s", target.Namespace, kube.LabelSelectorTableLines(target.PodSelector))
		rules := strings.Join(sourceRules, "\n")
		s.Prefix = []string{ruleType, targetString, rules}

		if len(target.Peers) == 0 {
			s.Append("no pods, no ips", "no ports, no protocols")
		} else {
			for _, peer := range target.Peers {
				s.Append(PeerMatcherTableLines(peer)...)
			}
		}
	}
}

func PeerMatcherTableLines(peer PeerMatcher) []string {
	switch a := peer.(type) {
	case *AllPeersMatcher:
		return []string{"all pods, all ips", "all ports, all protocols"}
	case *PortsForAllPeersMatcher:
		pps := PortMatcherTableLines(a.Port)
		return []string{"all pods, all ips", strings.Join(pps, "\n")}
	case *IPPeerMatcher:
		return IPPeerMatcherTableLines(a)
	case *PodPeerMatcher:
		return PodPeerMatcherTableLines(a)
	default:
		panic(errors.Errorf("invalid PeerMatcher type %T", a))
	}
}

func IPPeerMatcherTableLines(ip *IPPeerMatcher) []string {
	peer := ip.IPBlock.CIDR + "\n" + fmt.Sprintf("except %+v", ip.IPBlock.Except)
	pps := PortMatcherTableLines(ip.Port)
	return []string{peer, strings.Join(pps, "\n")}
}

func PodPeerMatcherTableLines(nsPodMatcher *PodPeerMatcher) []string {
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
	return []string{"namespace: " + namespaces + "\n" + "pods: " + pods, strings.Join(PortMatcherTableLines(nsPodMatcher.Port), "\n")}
}

func PortMatcherTableLines(pm PortMatcher) []string {
	switch port := pm.(type) {
	case *AllPortMatcher:
		return []string{"all ports, all protocols"}
	case *SpecificPortMatcher:
		var lines []string
		for _, portProtocol := range port.Ports {
			if portProtocol.Port == nil {
				lines = append(lines, "all ports on protocol "+string(portProtocol.Protocol))
			} else {
				lines = append(lines, "port "+portProtocol.Port.String()+" on protocol "+string(portProtocol.Protocol))
			}
		}
		for _, portRange := range port.PortRanges {
			lines = append(lines, fmt.Sprintf("ports [%d, %d] on protocol %s", portRange.From, portRange.To, portRange.Protocol))
		}
		return lines
	default:
		panic(errors.Errorf("invalid PortMatcher type %T", port))
	}
}
