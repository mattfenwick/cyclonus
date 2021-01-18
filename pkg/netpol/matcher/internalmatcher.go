package matcher

import (
	"encoding/json"
	"github.com/pkg/errors"
	"sort"
)

type InternalMatcher interface {
	Allows(peer *InternalPeer, portProtocol *PortProtocol) bool
}

func CombineInternalMatchers(a InternalMatcher, b InternalMatcher) InternalMatcher {
	switch l := a.(type) {
	case *AllInternalMatcher:
		return a
	case *NoneInternalMatcher:
		return b
	case *SpecificInternalMatcher:
		switch r := b.(type) {
		case *AllInternalMatcher:
			return b
		case *NoneInternalMatcher:
			return a
		case *SpecificInternalMatcher:
			for _, val := range r.NamespacePods {
				l.Add(val)
			}
			return l
		default:
			panic(errors.Errorf("invalid InternalMatcher type %T", b))
		}
	default:
		panic(errors.Errorf("invalid InternalMatcher type %T", a))
	}
}

// TODO is this possible, where only IPs are allowed?
//   maybe indirectly through: 1) deny all, 2) allow external with 0.0.0.0
type NoneInternalMatcher struct{}

func (n *NoneInternalMatcher) Allows(peer *InternalPeer, portProtocol *PortProtocol) bool {
	return false
}

func (n *NoneInternalMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "no internal",
	})
}

type AllInternalMatcher struct{}

func (a *AllInternalMatcher) Allows(peer *InternalPeer, portProtocol *PortProtocol) bool {
	return true
}

func (a *AllInternalMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "all internal",
	})
}

type SpecificInternalMatcher struct {
	NamespacePods map[string]*NamespacePodMatcher
}

func NewSpecificInternalMatcher(matchers ...*NamespacePodMatcher) *SpecificInternalMatcher {
	sim := &SpecificInternalMatcher{NamespacePods: map[string]*NamespacePodMatcher{}}
	for _, matcher := range matchers {
		sim.Add(matcher)
	}
	return sim
}

func (a *SpecificInternalMatcher) SortedNamespacePods() []*NamespacePodMatcher {
	var matchers []*NamespacePodMatcher
	for _, m := range a.NamespacePods {
		matchers = append(matchers, m)
	}
	sort.Slice(matchers, func(i, j int) bool {
		return matchers[i].PrimaryKey() < matchers[j].PrimaryKey()
	})
	return matchers
}

func (a *SpecificInternalMatcher) Allows(peer *InternalPeer, portProtocol *PortProtocol) bool {
	for _, podPeer := range a.NamespacePods {
		if podPeer.Allows(peer, portProtocol) {
			return true
		}
	}
	return false
}

func (a *SpecificInternalMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":          "specific internal",
		"NamespacePods": a.NamespacePods,
	})
}

func (a *SpecificInternalMatcher) Add(newMatcher *NamespacePodMatcher) {
	key := newMatcher.PrimaryKey()
	if oldMatcher, ok := a.NamespacePods[key]; ok {
		a.NamespacePods[key] = oldMatcher.Combine(newMatcher.Port)
	} else {
		a.NamespacePods[key] = newMatcher
	}
}
