package matcher

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/slices"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/exp/maps"
	"strings"
)

// Policy is the root type
type Policy struct {
	Ingress map[string]*Target
	Egress  map[string]*Target
}

func NewPolicy() *Policy {
	return &Policy{Ingress: map[string]*Target{}, Egress: map[string]*Target{}}
}

func NewPolicyWithTargets(ingress []*Target, egress []*Target) *Policy {
	np := NewPolicy()
	np.AddTargets(true, ingress)
	np.AddTargets(false, egress)
	return np
}

func (p *Policy) SortedTargets() ([]*Target, []*Target) {
	key := func(t *Target) string { return t.GetPrimaryKey() }
	ingress := slices.SortOn(key, maps.Values(p.Ingress))
	egress := slices.SortOn(key, maps.Values(p.Egress))
	return ingress, egress
}

func (p *Policy) AddTargets(isIngress bool, targets []*Target) {
	for _, target := range targets {
		p.AddTarget(isIngress, target)
	}
}

func (p *Policy) AddTarget(isIngress bool, target *Target) *Target {
	pk := target.GetPrimaryKey()
	var dict map[string]*Target
	if isIngress {
		dict = p.Ingress
	} else {
		dict = p.Egress
	}
	if prev, ok := dict[pk]; ok {
		combined := prev.Combine(target)
		dict[pk] = combined
	} else {
		dict[pk] = target
	}
	return dict[pk]
}

func (p *Policy) TargetsApplyingToPod(isIngress bool, namespace string, podLabels map[string]string) []*Target {
	var targets []*Target
	var dict map[string]*Target
	if isIngress {
		dict = p.Ingress
	} else {
		dict = p.Egress
	}
	for _, target := range dict {
		if target.IsMatch(namespace, podLabels) {
			targets = append(targets, target)
		}
	}
	return targets
}

type DirectionResult struct {
	AllowingTargets []*Target
	DenyingTargets  []*Target
}

func (d *DirectionResult) IsAllowed() bool {
	return len(d.AllowingTargets) > 0 || len(d.DenyingTargets) == 0
}

type AllowedResult struct {
	Ingress *DirectionResult
	Egress  *DirectionResult
}

func (ar *AllowedResult) Table() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Type", "Action", "Target"})

	addTargetsToTable(table, "Ingress", "Allow", ar.Ingress.AllowingTargets)
	addTargetsToTable(table, "Ingress", "Deny", ar.Ingress.DenyingTargets)
	table.Append([]string{"", "", ""})
	addTargetsToTable(table, "Egress", "Allow", ar.Egress.AllowingTargets)
	addTargetsToTable(table, "Egress", "Deny", ar.Egress.DenyingTargets)
	table.SetFooter([]string{"Is allowed?", fmt.Sprintf("%t", ar.IsAllowed()), ""})

	table.Render()
	return tableString.String()
}

func addTargetsToTable(table *tablewriter.Table, ruleType string, action string, targets []*Target) {
	sortedTargets := slices.SortOn(func(t *Target) string { return t.GetPrimaryKey() }, targets)
	for _, t := range sortedTargets {
		targetString := fmt.Sprintf("namespace: %s\n%s", t.Namespace, kube.LabelSelectorTableLines(t.PodSelector))
		table.Append([]string{ruleType, action, targetString})
	}
}

func (ar *AllowedResult) IsAllowed() bool {
	return ar.Ingress.IsAllowed() && ar.Egress.IsAllowed()
}

// IsTrafficAllowed returns:
// - whether the traffic is allowed
// - which rules allowed the traffic
// - which rules matched the traffic target
func (p *Policy) IsTrafficAllowed(traffic *Traffic) *AllowedResult {
	return &AllowedResult{
		Ingress: p.IsIngressOrEgressAllowed(traffic, true),
		Egress:  p.IsIngressOrEgressAllowed(traffic, false),
	}
}

func (p *Policy) IsIngressOrEgressAllowed(traffic *Traffic, isIngress bool) *DirectionResult {
	var target *TrafficPeer
	var peer *TrafficPeer
	if isIngress {
		target = traffic.Destination
		peer = traffic.Source
	} else {
		target = traffic.Source
		peer = traffic.Destination
	}

	// 1. if target is external to cluster -> allow
	//   this is because we can't stop external hosts from sending or receiving traffic
	if target.Internal == nil {
		return &DirectionResult{AllowingTargets: nil, DenyingTargets: nil}
	}

	matchingTargets := p.TargetsApplyingToPod(isIngress, target.Internal.Namespace, target.Internal.PodLabels)

	// 2. No targets match => automatic allow
	if len(matchingTargets) == 0 {
		return &DirectionResult{AllowingTargets: nil, DenyingTargets: nil}
	}

	// 3. Check if any matching targets allow this traffic
	pair := slices.Partition(func(t *Target) bool {
		return t.Allows(peer, traffic.ResolvedPort, traffic.ResolvedPortName, traffic.Protocol)
	}, matchingTargets)
	allowers, deniers := pair.Fst, pair.Snd

	return &DirectionResult{AllowingTargets: allowers, DenyingTargets: deniers}
}

func (p *Policy) Simplify() {
	for _, ingress := range p.Ingress {
		ingress.Simplify()
	}
	for _, egress := range p.Egress {
		egress.Simplify()
	}
}
