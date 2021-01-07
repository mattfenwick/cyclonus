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
		lines = append(lines, t.GetPrimaryKey())
		if len(t.SourceRules) != 0 {
			lines = append(lines, "  source rules:")
			for _, sr := range t.SourceRules {
				lines = append(lines, fmt.Sprintf("    %s/%s", sr.Namespace, sr.Name))
			}
		}
		switch a := t.Edge.(type) {
		case *NoneEdgeMatcher:
			lines = append(lines, "  all ingress blocked")
		case *EdgePeerPortMatcher:
			lines = append(lines, "  ingress:")
			lines = append(lines, ExplainEdgePeerPortMatcher(a)...)
		default:
			panic(errors.Errorf("invalid EdgeMatcher type %T", t.Edge))
		}

		lines = append(lines, "")
	}

	// 2. egress
	for _, t := range policies.Egress {
		lines = append(lines, t.GetPrimaryKey())
		if len(t.SourceRules) != 0 {
			lines = append(lines, "  source rules:")
			for _, sr := range t.SourceRules {
				lines = append(lines, fmt.Sprintf("    %s/%s", sr.Namespace, sr.Name))
			}
		}

		switch a := t.Edge.(type) {
		case *NoneEdgeMatcher:
			lines = append(lines, "  all egress blocked")
		case *EdgePeerPortMatcher:
			lines = append(lines, "  egress:")
			lines = append(lines, ExplainEdgePeerPortMatcher(a)...)
		default:
			panic(errors.Errorf("invalid EdgeMatcher type %T", t.Edge))
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
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
