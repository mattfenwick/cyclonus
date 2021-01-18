package matcher

import (
	"encoding/json"
	"github.com/pkg/errors"
)

type IPMatcher interface {
	Allows(ip string, portProtocol *PortProtocol) bool
}

func CombineIPMatchers(a IPMatcher, b IPMatcher) IPMatcher {
	switch l := a.(type) {
	case *AllIPMatcher:
		return a
	case *NoneIPMatcher:
		return b
	case *SpecificIPMatcher:
		switch r := b.(type) {
		case *AllIPMatcher:
			return b
		case *NoneIPMatcher:
			return a
		case *SpecificIPMatcher:
			return l.Combine(r)
		default:
			panic(errors.Errorf("invalid IPMatcher type %T", b))
		}
	default:
		panic(errors.Errorf("invalid IPMatcher type %T", a))
	}
}

type AllIPMatcher struct{}

func (aip *AllIPMatcher) Allows(ip string, portProtocol *PortProtocol) bool {
	return true
}

func (aip *AllIPMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "All IP",
	})
}

type NoneIPMatcher struct{}

func (aip *NoneIPMatcher) Allows(ip string, portProtocol *PortProtocol) bool {
	return false
}

func (aip *NoneIPMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "No IP",
	})
}

type SpecificIPMatcher struct {
	PortsForAllIPs PortMatcher
	IPBlocks       map[string]*IPBlockMatcher
}

func NewSpecificIPMatcher(portsForAllIPs PortMatcher, blocks ...*IPBlockMatcher) *SpecificIPMatcher {
	sip := &SpecificIPMatcher{
		PortsForAllIPs: portsForAllIPs,
		IPBlocks:       map[string]*IPBlockMatcher{},
	}
	for _, block := range blocks {
		sip.AddIPMatcher(block)
	}
	return sip
}

func (sip *SpecificIPMatcher) Allows(ip string, portProtocol *PortProtocol) bool {
	if sip.PortsForAllIPs.Allows(portProtocol.Port, portProtocol.Protocol) {
		return true
	}
	for _, ipMatcher := range sip.IPBlocks {
		if ipMatcher.Allows(ip, portProtocol) {
			return true
		}
	}
	return false
}

func (sip *SpecificIPMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":           "Specific IPs",
		"PortsForAllIPs": sip.PortsForAllIPs,
		"IPBlocks":       sip.IPBlocks,
	})
}

func (sip *SpecificIPMatcher) Combine(other *SpecificIPMatcher) *SpecificIPMatcher {
	ipMatchers := map[string]*IPBlockMatcher{}
	for key, ip := range sip.IPBlocks {
		ipMatchers[key] = ip
	}
	for key, ip := range other.IPBlocks {
		if matcher, ok := ipMatchers[key]; ok {
			ipMatchers[key] = matcher.Combine(ip)
		} else {
			ipMatchers[key] = ip
		}
	}
	return &SpecificIPMatcher{
		PortsForAllIPs: CombinePortMatchers(sip.PortsForAllIPs, other.PortsForAllIPs),
		IPBlocks:       ipMatchers}
}

func (sip *SpecificIPMatcher) AddIPMatcher(ip *IPBlockMatcher) {
	key := ip.PrimaryKey()
	if matcher, ok := sip.IPBlocks[key]; ok {
		sip.IPBlocks[key] = matcher.Combine(ip)
	} else {
		sip.IPBlocks[key] = ip
	}
}
