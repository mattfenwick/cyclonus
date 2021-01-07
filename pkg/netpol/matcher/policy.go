package matcher

// This is the root type
type Policy struct {
	Ingress map[string]*Target
	Egress  map[string]*Target
}

func NewPolicy() *Policy {
	return &Policy{Ingress: map[string]*Target{}, Egress: map[string]*Target{}}
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
		if target.Edge.Allows(peer, traffic.PortProtocol) {
			allowers = append(allowers, target)
		}
	}
	if len(allowers) > 0 {
		return &DirectionResult{IsAllowed: true, AllowingTargets: allowers, MatchingTargets: matchingTargets}
	}

	// 4. Otherwise, deny
	return &DirectionResult{IsAllowed: false, AllowingTargets: nil, MatchingTargets: matchingTargets}
}
