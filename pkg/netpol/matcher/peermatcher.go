package matcher

import "github.com/pkg/errors"

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

type AllPeerMatcher struct{}

func (aem *AllPeerMatcher) Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool {
	return true
}

type SpecificPeerMatcher struct {
	IP       map[string]*IPBlockMatcher
	Internal InternalMatcher
}

func (em *SpecificPeerMatcher) Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool {
	// can always match by ip
	for _, ipMatcher := range em.IP {
		if ipMatcher.Allows(peer.IP, portProtocol) {
			return true
		}
	}
	// internal? can also match by pod
	if !peer.IsExternal() {
		return em.Internal.Allows(peer.Internal, portProtocol)
	}

	return false
}

func (em *SpecificPeerMatcher) Combine(other *SpecificPeerMatcher) *SpecificPeerMatcher {
	ipMatchers := map[string]*IPBlockMatcher{}
	for key, ip := range em.IP {
		ipMatchers[key] = ip
	}
	for key, ip := range other.IP {
		if matcher, ok := ipMatchers[key]; ok {
			ipMatchers[key] = matcher.Combine(ip)
		} else {
			ipMatchers[key] = ip
		}
	}
	return &SpecificPeerMatcher{
		IP:       ipMatchers,
		Internal: CombineInternalMatchers(em.Internal, other.Internal),
	}
}

func (em *SpecificPeerMatcher) AddIPMatcher(ip *IPBlockMatcher) {
	key := ip.PrimaryKey()
	if matcher, ok := em.IP[key]; ok {
		em.IP[key] = matcher.Combine(ip)
	} else {
		em.IP[key] = ip
	}
}
