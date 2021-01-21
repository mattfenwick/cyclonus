package matcher

import (
	"encoding/json"
	"github.com/pkg/errors"
)

type PeerMatcher interface {
	Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool
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

func (nem *NonePeerMatcher) Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool {
	return false
}

func (nem *NonePeerMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "no peers",
	})
}

type AllPeerMatcher struct{}

func (aem *AllPeerMatcher) Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool {
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

func (em *SpecificPeerMatcher) Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool {
	// can always match by ip
	if em.IP.Allows(peer.IP, portProtocol) {
		return true
	}
	// internal? can also match by pod
	if !peer.IsExternal() {
		return em.Internal.Allows(peer.Internal, portProtocol)
	}

	return false
}

func (em *SpecificPeerMatcher) Combine(other *SpecificPeerMatcher) *SpecificPeerMatcher {
	return &SpecificPeerMatcher{
		IP:       CombineIPMatchers(em.IP, other.IP),
		Internal: CombineInternalMatchers(em.Internal, other.Internal),
	}
}
