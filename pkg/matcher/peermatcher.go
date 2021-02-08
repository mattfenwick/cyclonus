package matcher

import (
	"encoding/json"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

type PeerMatcher interface {
	Allows(peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool
}

func CombinePeerMatchers(a PeerMatcher, b PeerMatcher) PeerMatcher {
	switch l := a.(type) {
	case *NonePeerMatcher:
		return b
	case *AllPeerMatcher:
		return a
	case *SpecificPeerMatcher:
		switch r := b.(type) {
		case *NonePeerMatcher:
			return a
		case *AllPeerMatcher:
			return b
		case *SpecificPeerMatcher:
			return l.Combine(r)
		default:
			panic(errors.Errorf("invalid PeerMatcher type %T", b))
		}
	default:
		panic(errors.Errorf("invalid PeerMatcher type %T", a))
	}
}

type NonePeerMatcher struct{}

func (nem *NonePeerMatcher) Allows(peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool {
	return false
}

func (nem *NonePeerMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "no peers",
	})
}

type AllPeerMatcher struct{}

func (aem *AllPeerMatcher) Allows(peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool {
	return true
}

func (aem *AllPeerMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "all peers",
	})
}

type SpecificPeerMatcher struct {
	IP       IPMatcher
	Internal InternalMatcher
}

func (em *SpecificPeerMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":     "specific peers",
		"IP":       em.IP,
		"Internal": em.Internal,
	})
}

func (em *SpecificPeerMatcher) Allows(peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool {
	// can always match by ip
	if em.IP.Allows(peer.IP, portInt, portName, protocol) {
		return true
	}
	// internal? can also match by pod
	if !peer.IsExternal() {
		return em.Internal.Allows(peer.Internal, portInt, portName, protocol)
	}

	return false
}

func (em *SpecificPeerMatcher) Combine(other *SpecificPeerMatcher) *SpecificPeerMatcher {
	return &SpecificPeerMatcher{
		IP:       CombineIPMatchers(em.IP, other.IP),
		Internal: CombineInternalMatchers(em.Internal, other.Internal),
	}
}
