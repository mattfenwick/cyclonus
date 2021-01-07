package netpol

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func nodeHelp(node Node, depth int, f func(Node, int)) {
	f(node, depth)
	for _, c := range node.Children() {
		nodeHelp(c, depth+1, f)
	}
}

func NodeTraverse(node Node, f func(Node, int)) {
	nodeHelp(node, 0, f)
}

func NodePrettyPrint(rootNode Node) string {
	var lines []string
	NodeTraverse(rootNode, func(node Node, i int) {
		prefix := strings.Repeat(" ", i)
		lines = append(lines, prefix+node.Print())
	})
	return strings.Join(lines, "\n")
}

type Node interface {
	Children() []Node
	Print() string
}

type Branch struct {
	Operation string
	Nodes     []Node
}

func (b *Branch) Children() []Node {
	return b.Nodes
}

func (b *Branch) Print() string {
	return b.Operation
}

type MatchKeyValue struct {
	Key   string
	Value string
}

func (mkv *MatchKeyValue) Children() []Node {
	return nil
}

func (mkv *MatchKeyValue) Print() string {
	return fmt.Sprintf("MATCH: %s: %s", mkv.Key, mkv.Value)
}

type MatchExpression metav1.LabelSelectorRequirement

func (me *MatchExpression) Children() []Node {
	return nil
}

func (me *MatchExpression) Print() string {
	return fmt.Sprintf("MATCH-EXPRESSION: %s: %+v, %s", me.Key, me.Values, me.Operator)
}

type Leaf struct {
	Value string
}

func (l *Leaf) Children() []Node {
	return nil
}

func (l *Leaf) Print() string {
	return l.Value
}

type Port struct {
	Protocol        string
	PortOrNamedPort string
}

func (p *Port) Children() []Node {
	return nil
}

func (p *Port) Print() string {
	return fmt.Sprintf("Port: %s: %s", p.Protocol, p.PortOrNamedPort)
}

type NamespaceSelector struct {
	CurrentNamespace bool
	AllNamespaces    bool
	Selector         Node
}

func (n *NamespaceSelector) Children() []Node {
	if n.Selector == nil {
		return []Node{}
	}
	return []Node{n.Selector}
}

func (n *NamespaceSelector) Print() string {
	if n.CurrentNamespace {
		return "current namespace"
	} else if n.AllNamespaces {
		return "all namespaces"
	}
	// TODO not sure what to do here
	return fmt.Sprintf("Namespace label selector: %+v", n.Selector)
}

func Reduce(policy *networkingv1.NetworkPolicy) Node {
	targetSelector := ReduceSelector(policy.Spec.PodSelector)
	nodes := []Node{targetSelector}

	isIngress, isEgress := false, false
	for _, pType := range policy.Spec.PolicyTypes {
		switch pType {
		case networkingv1.PolicyTypeIngress:
			isIngress = true
		case networkingv1.PolicyTypeEgress:
			isEgress = true
		}
	}
	if isEgress {
		egress := ReduceEgresses(policy.Spec.Egress)
		nodes = append(nodes, egress)
	}
	if isIngress {
		ingress := ReduceIngresses(policy.Spec.Ingress)
		nodes = append(nodes, ingress)
	}

	return &Branch{
		Operation: "&&",
		Nodes:     nodes,
	}
}

func ReduceSelector(sel metav1.LabelSelector) Node {
	return &Branch{
		Operation: "label selector: &&",
		Nodes: []Node{
			ReduceMatchLabels(sel.MatchLabels),
			ReduceMatchExpressions(sel.MatchExpressions),
		},
	}
}

func ReduceMatchLabels(labels map[string]string) Node {
	var nodes []Node
	for key, val := range labels {
		nodes = append(nodes, &MatchKeyValue{Key: key, Value: val})
	}
	return &Branch{
		Operation: "match labels: &&",
		Nodes:     nodes,
	}
}

func ReduceMatchExpressions(exps []metav1.LabelSelectorRequirement) Node {
	var nodes []Node
	for _, e := range exps {
		me := MatchExpression(e)
		nodes = append(nodes, &me)
	}
	return &Branch{
		Operation: "match expressions: &&",
		Nodes:     nodes,
	}
}

func ReducePorts(ports []networkingv1.NetworkPolicyPort) Node {
	var nodes []Node
	for _, p := range ports {
		protocol := v1.ProtocolTCP
		if p.Protocol != nil {
			protocol = *p.Protocol
		}
		nodes = append(nodes, &Port{
			Protocol:        string(protocol),
			PortOrNamedPort: p.Port.String(),
		})
	}
	return &Branch{
		Operation: "Ports: ||",
		Nodes:     nodes,
	}
}

func ReduceNamespaceSelector(sel *metav1.LabelSelector) *NamespaceSelector {
	if sel == nil {
		return &NamespaceSelector{CurrentNamespace: true}
	} else if len(sel.MatchLabels) == 0 {
		return &NamespaceSelector{AllNamespaces: true}
	}
	return &NamespaceSelector{Selector: ReduceSelector(*sel)}
}

func ReducePodSelector(sel *metav1.LabelSelector) Node {
	if sel == nil {
		return &Leaf{Value: "all pods"}
	}
	return &Branch{Operation: "", Nodes: []Node{ReduceSelector(*sel)}}
}

func ReduceIpBlock(ipBlock *networkingv1.IPBlock) Node {
	return &Leaf{Value: fmt.Sprintf("IPBlock: %s except %+v", ipBlock.CIDR, ipBlock.Except)}
}

func ReduceNetworkPolicyPeer(isEgress bool, npp networkingv1.NetworkPolicyPeer) Node {
	operation := "Egress to: &&"
	if !isEgress {
		operation = "Ingress from: &&"
	}
	if npp.IPBlock != nil && (npp.PodSelector != nil || npp.NamespaceSelector != nil) {
		panic("invalid NetworkPolicyPeer -- IPBlock not nil, along with PodSelector or NamespaceSelector")
	}
	if npp.IPBlock == nil && npp.PodSelector == nil && npp.NamespaceSelector == nil {
		panic("invalid NetworkPolicyPeer -- all nil")
	}
	var nodes []Node
	if npp.IPBlock != nil {
		nodes = append(nodes, ReduceIpBlock(npp.IPBlock))
	} else {
		nodes = append(nodes, ReducePodSelector(npp.PodSelector), ReduceNamespaceSelector(npp.NamespaceSelector))
	}
	return &Branch{
		Operation: operation,
		Nodes:     nodes,
	}
}

func ReduceEgress(egress networkingv1.NetworkPolicyEgressRule) Node {
	tos := &Branch{
		Operation: "Egress tos: ||",
	}
	for _, to := range egress.To {
		tos.Nodes = append(tos.Nodes, ReduceNetworkPolicyPeer(true, to))
	}

	return &Branch{
		Operation: "Egress: &&",
		Nodes:     []Node{ReducePorts(egress.Ports), tos},
	}
}

func ReduceEgresses(egresses []networkingv1.NetworkPolicyEgressRule) Node {
	var nodes []Node
	for _, egress := range egresses {
		nodes = append(nodes, ReduceEgress(egress))
	}
	return &Branch{
		Operation: "Egresses ||",
		Nodes:     nodes,
	}
}

func ReduceIngress(ingress networkingv1.NetworkPolicyIngressRule) Node {
	froms := &Branch{
		Operation: "Ingress froms: ||",
	}
	for _, from := range ingress.From {
		froms.Nodes = append(froms.Nodes, ReduceNetworkPolicyPeer(false, from))
	}

	return &Branch{
		Operation: "Ingress: &&",
		Nodes:     []Node{ReducePorts(ingress.Ports), froms},
	}
}

func ReduceIngresses(ingresses []networkingv1.NetworkPolicyIngressRule) Node {
	var nodes []Node
	for _, ingress := range ingresses {
		nodes = append(nodes, ReduceIngress(ingress))
	}
	return &Branch{
		Operation: "Ingresses &&",
		Nodes:     nodes,
	}
}
