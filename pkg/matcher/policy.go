package matcher

import "sort"

// This is the root type
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

func (np *Policy) SortedTargets() ([]*Target, []*Target) {
	var ingress, egress []*Target
	for _, rule := range np.Ingress {
		ingress = append(ingress, rule)
	}
	sort.Slice(ingress, func(i, j int) bool {
		return ingress[i].GetPrimaryKey() < ingress[j].GetPrimaryKey()
	})
	for _, rule := range np.Egress {
		egress = append(egress, rule)
	}
	sort.Slice(egress, func(i, j int) bool {
		return egress[i].GetPrimaryKey() < egress[j].GetPrimaryKey()
	})
	return ingress, egress
}

func (np *Policy) AddTargets(isIngress bool, targets []*Target) {
	for _, target := range targets {
		np.AddTarget(isIngress, target)
	}
}

func (np *Policy) AddTarget(isIngress bool, target *Target) *Target {
	pk := target.GetPrimaryKey()
	var dict map[string]*Target
	if isIngress {
		dict = np.Ingress
	} else {
		dict = np.Egress
	}
	if prev, ok := dict[pk]; ok {
		combined := prev.Combine(target)
		dict[pk] = combined
	} else {
		dict[pk] = target
	}
	return dict[pk]
}

func (np *Policy) TargetsApplyingToPod(isIngress bool, namespace string, podLabels map[string]string) []*Target {
	var targets []*Target
	var dict map[string]*Target
	if isIngress {
		dict = np.Ingress
	} else {
		dict = np.Egress
	}
	for _, target := range dict {
		if target.IsMatch(namespace, podLabels) {
			targets = append(targets, target)
		}
	}
	return targets
}

type DirectionResult struct {
	IsAllowed       bool
	AllowingTargets []*Target
	MatchingTargets []*Target
}

type AllowedResult struct {
	Ingress *DirectionResult
	Egress  *DirectionResult
}

func (ar *AllowedResult) IsAllowed() bool {
	return ar.Ingress.IsAllowed && ar.Egress.IsAllowed
}

// IsTrafficAllowed returns:
// - whether the traffic is allowed
// - which rules allowed the traffic
// - which rules matched the traffic target
func (np *Policy) IsTrafficAllowed(traffic *Traffic) *AllowedResult {
	return &AllowedResult{
		Ingress: np.IsIngressOrEgressAllowed(traffic, true),
		Egress:  np.IsIngressOrEgressAllowed(traffic, false),
	}
}

func (np *Policy) IsIngressOrEgressAllowed(traffic *Traffic, isIngress bool) *DirectionResult {
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
		return &DirectionResult{IsAllowed: true, AllowingTargets: nil, MatchingTargets: nil}
	}

	matchingTargets := np.TargetsApplyingToPod(isIngress, target.Internal.Namespace, target.Internal.PodLabels)

	// 2. No targets match => automatic allow
	if len(matchingTargets) == 0 {
		return &DirectionResult{IsAllowed: true, AllowingTargets: nil, MatchingTargets: nil}
	}

	// 3. Check if any matching targets allow this traffic
	var allowers []*Target
	for _, target := range matchingTargets {
		if target.Peer.Allows(peer, traffic.PortProtocol) {
			allowers = append(allowers, target)
		}
	}
	if len(allowers) > 0 {
		return &DirectionResult{IsAllowed: true, AllowingTargets: allowers, MatchingTargets: matchingTargets}
	}

	// 4. Otherwise, deny
	return &DirectionResult{IsAllowed: false, AllowingTargets: nil, MatchingTargets: matchingTargets}
}
