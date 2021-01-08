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
	switch a := target.Edge.(type) {
	case *NoneEdgeMatcher:
		lines = append(lines, fmt.Sprintf("  all %s blocked", targetType))
	case *EdgePeerPortMatcher:
		lines = append(lines, fmt.Sprintf("  %s:", targetType))
		lines = append(lines, ExplainEdgePeerPortMatcher(a)...)
	default:
		panic(errors.Errorf("invalid EdgeMatcher type %T", target.Edge))
	}

	lines = append(lines, "")
	return lines
}

func ExplainEdgePeerPortMatcher(tp *EdgePeerPortMatcher) []string {
	var lines []string
	for _, sd := range tp.Matchers {
		var sourceDest, port string
		switch t := sd.Peer.(type) {
		case *MatchingPodsInAllNamespacesPeerMatcher:
			sourceDest = fmt.Sprintf("pods matching %s in all namespaces",
				SerializeLabelSelector(t.PodSelector))
		case *MatchingPodsInMatchingNamespacesPeerMatcher:
			sourceDest = fmt.Sprintf("pods matching %s in namespaces matching %s",
				SerializeLabelSelector(t.PodSelector),
				SerializeLabelSelector(t.NamespaceSelector))
		case *AllPodsInMatchingNamespacesPeerMatcher:
			sourceDest = fmt.Sprintf("all pods in namespaces matching %s",
				SerializeLabelSelector(t.NamespaceSelector))
		case *AllPodsInPolicyNamespacePeerMatcher:
			sourceDest = fmt.Sprintf("all pods in namespace %s", t.Namespace)
		case *MatchingPodsInPolicyNamespacePeerMatcher:
			sourceDest = fmt.Sprintf("pods matching %s in namespace %s",
				SerializeLabelSelector(t.PodSelector), t.Namespace)
		case *AllPodsAllNamespacesPeerMatcher:
			sourceDest = "all pods in all namespaces"
		case *AnywherePeerMatcher:
			sourceDest = "anywhere: all pods in all namespaces and all IPs"
		case *IPBlockPeerMatcher:
			sourceDest = fmt.Sprintf("IPBlock: cidr %s, except %+v", t.IPBlock.CIDR, t.IPBlock.Except)
		default:
			panic(errors.Errorf("unexpected PeerMatcher type %T", t))
		}
		switch p := sd.Port.(type) {
		case *AllPortsOnProtocolMatcher:
			port = fmt.Sprintf("all ports on protocol %s", p.Protocol)
		case *AllPortsAllProtocolsMatcher:
			port = "all ports all protocols"
		case *ExactPortProtocolMatcher:
			port = fmt.Sprintf("port %s on protocol %s", p.Port.String(), p.Protocol)
		default:
			panic(errors.Errorf("unexpected Port type %T", p))
		}
		lines = append(lines, "  - "+sourceDest, "    "+port)
	}

	return lines
}
