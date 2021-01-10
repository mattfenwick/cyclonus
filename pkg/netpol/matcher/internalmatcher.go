package matcher

import (
	"encoding/json"
	"github.com/pkg/errors"
)

type InternalMatcher interface {
	Allows(peer *InternalPeer, portProtocol *PortProtocol) bool
}

func CombineInternalMatchers(a InternalMatcher, b InternalMatcher) InternalMatcher {
	switch l := a.(type) {
	case *NoneInternalMatcher:
		return b
	case *AllInternalMatcher:
		return a
	case *SpecificInternalMatcher:
		switch r := b.(type) {
		case *NoneInternalMatcher:
			return a
		case *AllInternalMatcher:
			return b
		case *SpecificInternalMatcher:
			for _, val := range r.Pods {
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
	Pods map[string]*NamespacePodMatcher
}

func (a *SpecificInternalMatcher) Allows(peer *InternalPeer, portProtocol *PortProtocol) bool {
	for _, podPeer := range a.Pods {
		if podPeer.Allows(peer, portProtocol) {
			return true
		}
	}
	return false
}

func (a *SpecificInternalMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "specific internal",
		"Pods": a.Pods,
	})
}

func (a *SpecificInternalMatcher) Add(newMatcher *NamespacePodMatcher) {
	key := newMatcher.PrimaryKey()
	if oldMatcher, ok := a.Pods[key]; ok {
		a.Pods[key] = oldMatcher.Combine(newMatcher.Port)
	} else {
		a.Pods[key] = newMatcher
	}
}
