package matcher

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

// Target represents rules applied to a group of pods defined by a selector
type Target struct {
	Selector    *Selector
	Peers       []PeerMatcher
	SourceRules []*networkingv1.NetworkPolicy
	primaryKey  string
}

func (t *Target) String() string {
	return t.GetPrimaryKey()
}

func (t *Target) AppliesToNamesAndLabels(podNamesAndLabels *SelectorTargetPod) bool {
	return t.Selector.Matches(podNamesAndLabels)
}

func (t *Target) Allows(peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool {
	for _, peerMatcher := range t.Peers {
		if peerMatcher.Allows(peer, portInt, portName, protocol) {
			return true
		}
	}
	return false
}

// Combine creates a new Target combining the egress and ingress rules
// of the two original targets.  Neither input is modified.
// The Primary Keys of the two targets must match.
func (t *Target) Combine(other *Target) *Target {
	myPk := t.GetPrimaryKey()
	otherPk := other.GetPrimaryKey()
	if myPk != otherPk {
		panic(errors.Errorf("cannot combine targets: primary keys differ -- '%s' vs '%s'", myPk, otherPk))
	}

	return &Target{
		Selector:    t.Selector,
		Peers:       append(t.Peers, other.Peers...),
		SourceRules: append(t.SourceRules, other.SourceRules...),
	}
}

// GetPrimaryKey returns a deterministic combination of namespace and pod selectors
func (t *Target) GetPrimaryKey() string {
	return t.Selector.GetPrimaryKey()
}

// CombineTargetsIgnoringPrimaryKey creates a new target from the given namespace and pod selector,
// and combines all the edges and source rules from the original targets into the new target.
func CombineTargetsIgnoringPrimaryKey(selector *Selector, targets []*Target) *Target {
	if len(targets) == 0 {
		return nil
	}
	target := &Target{
		Selector:    selector,
		Peers:       targets[0].Peers,
		SourceRules: targets[0].SourceRules,
	}
	for _, t := range targets[1:] {
		target.Peers = append(target.Peers, t.Peers...)
		target.SourceRules = append(target.SourceRules, t.SourceRules...)
	}
	return target
}

func (t *Target) Simplify() {
	t.Peers = Simplify(t.Peers)
}
