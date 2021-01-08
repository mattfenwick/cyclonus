package matcher

import (
	"encoding/json"
	"github.com/pkg/errors"
)

type EdgeMatcher interface {
	Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool
}

func CombineEdgeMatchers(a EdgeMatcher, b EdgeMatcher) EdgeMatcher {
	switch this := a.(type) {
	case *EdgePeerPortMatcher:
		switch that := b.(type) {
		case *EdgePeerPortMatcher:
			return &EdgePeerPortMatcher{Matchers: append(this.Matchers, that.Matchers...)}
		//case *AllEdgeMatcher:
		//	return b
		case *NoneEdgeMatcher:
			return a
		default:
			panic(errors.Errorf("invalid EdgeMatcher type %T", b))
		}
	//case *AllEdgeMatcher:
	//	return a
	case *NoneEdgeMatcher:
		return b
	default:
		panic(errors.Errorf("invalid EdgeMatcher type %T", a))
	}
}

type EdgePeerPortMatcher struct {
	// TODO change this to:
	//   SourceDests map[string]*PeerPortMatcher
	//   where the key is the PK of PeerPortMatcher
	//   and add a (tp *EdgeMatcher)Add(sdap *PeerMatcher, port Port) method or something
	//   goal: nest ports under PeerMatchers
	Matchers []*PeerPortMatcher
}

func (eppm *EdgePeerPortMatcher) Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool {
	if len(eppm.Matchers) == 0 {
		panic(errors.Errorf("cannot have 0 matchers -- use NoneEdgeMatcher instead"))
	}
	for _, sd := range eppm.Matchers {
		if sd.Allows(peer, portProtocol) {
			return true
		}
	}
	return false
}

func (eppm *EdgePeerPortMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":     "edge",
		"Matchers": eppm.Matchers,
	})
}

// TODO is this necessary, or is it handled by AnywherePeerMatcher ?
//type AllEdgeMatcher struct{}
//
//func (aem *AllEdgeMatcher) Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool {
//	return true
//}

type NoneEdgeMatcher struct{}

func (nem *NoneEdgeMatcher) Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool {
	return false
}

func (nem *NoneEdgeMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "none",
	})
}
