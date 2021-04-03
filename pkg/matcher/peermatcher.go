package matcher

import (
	"encoding/json"
	v1 "k8s.io/api/core/v1"
)

var (
	NoPeers       = &NonePeerMatcher{}
	AllPeersPorts = &AllPeersMatcher{}
)

type PeerMatcher interface {
	Allows(peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool
}

type AllPeersMatcher struct{}

func (a *AllPeersMatcher) Allows(peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool {
	return true
}

func (a *AllPeersMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "all peers",
	})
}

type PortsForAllPeersMatcher struct {
	Port PortMatcher
}

func (a *PortsForAllPeersMatcher) Allows(peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool {
	return a.Port.Allows(portInt, portName, protocol)
}

func (a *PortsForAllPeersMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "all peers for port",
		"Port": a.Port,
	})
}

// TODO this type seems dumb and we should get rid of it, it's just awkward to carry around
type NonePeerMatcher struct{}

func (nem *NonePeerMatcher) Allows(peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool {
	return false
}

func (nem *NonePeerMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "no peers",
	})
}
