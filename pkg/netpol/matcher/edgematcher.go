package matcher

import "github.com/pkg/errors"

type EdgeMatcher interface {
	Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool
}

func CombineEdgeMatchers(a EdgeMatcher, b EdgeMatcher) EdgeMatcher {
	switch l := a.(type) {
	case *NoneEdgeMatcher:
		return b
	case *AllEdgeMatcher:
		return a
	case *SpecificEdgeMatcher:
		switch r := b.(type) {
		case *NoneEdgeMatcher:
			return a
		case *AllEdgeMatcher:
			return b
		case *SpecificEdgeMatcher:
			return l.Combine(r)
		default:
			panic(errors.Errorf("invalid EdgeMatcher type %T", b))
		}
	default:
		panic(errors.Errorf("invalid EdgeMatcher type %T", a))
	}
}

type NoneEdgeMatcher struct{}

func (nem *NoneEdgeMatcher) Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool {
	return false
}

type AllEdgeMatcher struct{}

func (aem *AllEdgeMatcher) Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool {
	return true
}

type SpecificEdgeMatcher struct {
	IP       map[string]*IPBlockPeerMatcher
	Internal InternalMatcher
}

func (em *SpecificEdgeMatcher) Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool {
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

func (em *SpecificEdgeMatcher) Combine(other *SpecificEdgeMatcher) *SpecificEdgeMatcher {
	ipMatchers := map[string]*IPBlockPeerMatcher{}
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
	return &SpecificEdgeMatcher{
		IP:       ipMatchers,
		Internal: CombineInternalMatchers(em.Internal, other.Internal),
	}
}

func (em *SpecificEdgeMatcher) AddIPMatcher(ip *IPBlockPeerMatcher) {
	key := ip.PrimaryKey()
	if matcher, ok := em.IP[key]; ok {
		em.IP[key] = matcher.Combine(ip)
	} else {
		em.IP[key] = ip
	}
}
