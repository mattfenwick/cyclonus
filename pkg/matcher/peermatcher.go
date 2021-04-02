package matcher

import (
	"encoding/json"
	v1 "k8s.io/api/core/v1"
)

type PeerMatcher interface {
	Allows(peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool
}

type AllPeerMatcher struct {
	Port PortMatcher
}

func (a *AllPeerMatcher) Allows(peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool {
	return a.Port.Allows(portInt, portName, protocol)
}

func (a *AllPeerMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "all peers",
		"Port": a.Port,
	})
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

//func CombinePeerMatchers(a PeerMatcher, b PeerMatcher) PeerMatcher {
//	switch l := a.(type) {
//	case *NonePeerMatcher:
//		return b
//	case *AllPeerMatcher:
//		return a
//	case *SpecificPeerMatcher:
//		switch r := b.(type) {
//		case *NonePeerMatcher:
//			return a
//		case *AllPeerMatcher:
//			return b
//		case *SpecificPeerMatcher:
//			return l.Combine(r)
//		default:
//			panic(errors.Errorf("invalid PeerMatcher type %T", b))
//		}
//	default:
//		panic(errors.Errorf("invalid PeerMatcher type %T", a))
//	}
//}

// TODO use this for consolidating/combining IPBlockMatchers
//func (sip *SpecificIPMatcher) Combine(other *SpecificIPMatcher) *SpecificIPMatcher {
//	ipMatchers := map[string]*IPPeerMatcher{}
//	for key, ip := range sip.IPBlocks {
//		ipMatchers[key] = ip
//	}
//	for key, ip := range other.IPBlocks {
//		if matcher, ok := ipMatchers[key]; ok {
//			ipMatchers[key] = matcher.Combine(ip)
//		} else {
//			ipMatchers[key] = ip
//		}
//	}
//	return &SpecificIPMatcher{
//		PortsForAllIPs: CombinePortMatchers(sip.PortsForAllIPs, other.PortsForAllIPs),
//		IPBlocks:       ipMatchers}
//}

//func (a *SpecificInternalMatcher) SortedNamespacePods() []*PodPeerMatcher {
//	var matchers []*PodPeerMatcher
//	for _, m := range a.NamespacePods {
//		matchers = append(matchers, m)
//	}
//	sort.Slice(matchers, func(i, j int) bool {
//		return matchers[i].PrimaryKey() < matchers[j].PrimaryKey()
//	})
//	return matchers
//}

//func CombinePortMatchers(a PortMatcher, b PortMatcher) PortMatcher {
//	switch l := a.(type) {
//	case *AllPortMatcher:
//		return a
//	case *NonePortMatcher:
//		return b
//	case *SpecificPortMatcher:
//		switch r := b.(type) {
//		case *AllPortMatcher:
//			return b
//		case *NonePortMatcher:
//			return a
//		case *SpecificPortMatcher:
//			return l.Combine(r)
//		default:
//			panic(errors.Errorf("invalid Port type %T", b))
//		}
//	default:
//		panic(errors.Errorf("invalid Port type %T", a))
//	}
//}

//func (ppm *PodPeerMatcher) Combine(otherPort PortMatcher) *PodPeerMatcher {
//	return &PodPeerMatcher{
//		Namespace: ppm.Namespace,
//		Pod:       ppm.Pod,
//		Port:      CombinePortMatchers(ppm.Port, otherPort),
//	}
//}

//func (i *IPPeerMatcher) Combine(other *IPPeerMatcher) *IPPeerMatcher {
//	if i.PrimaryKey() != other.PrimaryKey() {
//		panic(errors.Errorf("unable to combine IPPeerMatcher values with different primary keys: %s vs %s", i.PrimaryKey(), other.PrimaryKey()))
//	}
//	return &IPPeerMatcher{
//		IPBlock: i.IPBlock,
//		Port:    CombinePortMatchers(i.Port, other.Port),
//	}
//}
