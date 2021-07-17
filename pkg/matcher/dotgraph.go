package matcher

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

type Graph struct {
	Nodes map[string]bool
	Edges []*Edge
}

func NewGraph() *Graph {
	return &Graph{Nodes: map[string]bool{}}
}

func (g *Graph) AddNode(node string) {
	g.Nodes[node] = true
}

func (g *Graph) AddEdge(from string, to string) {
	g.Edges = append(g.Edges, &Edge{From: from, To: to})
}

type Edge struct {
	From string
	To   string
}

func (g *Graph) Serialize() string {
	lines := []string{`digraph "netpols" {`}
	for node := range g.Nodes {
		lines = append(lines, fmt.Sprintf(`  "%s" [color=%s,penwidth=5];`, node, "red"))
	}
	for _, edge := range g.Edges {
		lines = append(lines, fmt.Sprintf(`  "%s" -> "%s";`, edge.From, edge.To))
	}
	return strings.Join(append(lines, "}"), "\n")
}

func BuildGraph(np *Policy) string {
	graph := NewGraph()
	// TODO add egress
	for _, rule := range np.Ingress {
		name := rule.Namespace + "/" + LabelSelectorGraph(rule.PodSelector)
		graph.AddNode(name)
		for _, peer := range rule.Peers {
			from := PeerMatcherGraph(peer)
			graph.AddEdge(from, name)
		}
	}
	return graph.Serialize()
}

func PeerMatcherGraph(peer PeerMatcher) string {
	switch a := peer.(type) {
	case *AllPeersMatcher:
		return "any"
	case *PortsForAllPeersMatcher:
		return "all pod/ip " + PortMatcherGraph(a.Port)
	case *IPPeerMatcher:
		return IPPeerMatcherGraph(a)
	case *PodPeerMatcher:
		return PodPeerMatcherGraph(a)
	default:
		panic(errors.Errorf("invalid PeerMatcher type %T", a))
	}
}

func IPPeerMatcherGraph(ip *IPPeerMatcher) string {
	peer := fmt.Sprintf("%s except %+v", ip.IPBlock.CIDR, ip.IPBlock.Except)
	return fmt.Sprintf("%s@%s", peer, PortMatcherGraph(ip.Port))
}

func PodPeerMatcherGraph(nsPodMatcher *PodPeerMatcher) string {
	var namespaces string
	switch ns := nsPodMatcher.Namespace.(type) {
	case *AllNamespaceMatcher:
		namespaces = "all"
	case *LabelSelectorNamespaceMatcher:
		namespaces = LabelSelectorGraph(ns.Selector)
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
		pods = LabelSelectorGraph(p.Selector)
	default:
		panic(errors.Errorf("invalid PodMatcher type %T", p))
	}
	return fmt.Sprintf("ns: %s; pod: %s; port: %s", namespaces, pods, PortMatcherGraph(nsPodMatcher.Port))
}

func LabelSelectorGraph(selector metav1.LabelSelector) string {
	if kube.IsLabelSelectorEmpty(selector) {
		return "all"
	}
	var kvs []string
	if len(selector.MatchLabels) > 0 {
		for key, val := range selector.MatchLabels {
			kvs = append(kvs, fmt.Sprintf("[%s: %s]", key, val))
		}
	}
	var exps []string
	if len(selector.MatchExpressions) > 0 {
		for _, exp := range selector.MatchExpressions {
			exps = append(exps, fmt.Sprintf("(%s %s %+v)", exp.Key, exp.Operator, exp.Values))
		}
	}
	return strings.Join(append(kvs, exps...), ", ")
}

func PortMatcherGraph(pm PortMatcher) string {
	switch port := pm.(type) {
	case *AllPortMatcher:
		return "any"
	case *SpecificPortMatcher:
		var pps []string
		for _, portProtocol := range port.Ports {
			if portProtocol.Port == nil {
				pps = append(pps, "any/"+string(portProtocol.Protocol))
			} else {
				pps = append(pps, fmt.Sprintf("%s/%s", portProtocol.Port.String(), string(portProtocol.Protocol)))
			}
		}
		for _, portRange := range port.PortRanges {
			pps = append(pps, fmt.Sprintf("[%d-%d]/%s", portRange.From, portRange.To, portRange.Protocol))
		}
		return strings.Join(pps, ", ")
	default:
		panic(errors.Errorf("invalid PortMatcher type %T", port))
	}
}
