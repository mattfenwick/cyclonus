package matcher

import (
	"encoding/json"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sort"
)

type PortMatcher interface {
	Allows(portInt int, portName string, protocol v1.Protocol) bool
}

// TODO add port range

type AllPortMatcher struct{}

func (ap *AllPortMatcher) Allows(portInt int, portName string, protocol v1.Protocol) bool {
	return true
}

func (ap *AllPortMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "all ports",
	})
}

// PortProtocolMatcher models a specific combination of port+protocol.  If port is nil,
// all ports are matched.
type PortProtocolMatcher struct {
	Port     *intstr.IntOrString
	Protocol v1.Protocol
}

// AllowsPort does not implement the PortMatcher interface, purposely!
func (p *PortProtocolMatcher) AllowsPort(portInt int, portName string, protocol v1.Protocol) bool {
	if p.Port != nil {
		return isPortMatch(*p.Port, portInt, portName) && p.Protocol == protocol
	}
	return p.Protocol == protocol
}

func (p *PortProtocolMatcher) Equals(other *PortProtocolMatcher) bool {
	if p.Protocol != other.Protocol {
		return false
	}
	if p.Port == nil && other.Port == nil {
		return true
	}
	if (p.Port == nil && other.Port != nil) || (p.Port != nil && other.Port == nil) {
		return false
	}
	return isIntStringEqual(*p.Port, *other.Port)
}

// SpecificPortMatcher models the case where traffic must match a named or numbered port
type SpecificPortMatcher struct {
	Ports []*PortProtocolMatcher
}

func (s *SpecificPortMatcher) Allows(portInt int, portName string, protocol v1.Protocol) bool {
	for _, matcher := range s.Ports {
		if matcher.AllowsPort(portInt, portName, protocol) {
			return true
		}
	}
	return false
}

func (s *SpecificPortMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":  "specific ports",
		"Ports": s.Ports,
	})
}

func (s *SpecificPortMatcher) Combine(other *SpecificPortMatcher) *SpecificPortMatcher {
	var pps []*PortProtocolMatcher
	for _, pp := range s.Ports {
		pps = append(pps, pp)
	}
	for _, otherPP := range other.Ports {
		for _, pp := range pps {
			if pp.Equals(otherPP) {
				break
			}
			pps = append(pps, otherPP)
		}
	}
	sort.Slice(pps, func(i, j int) bool {
		// first, run it forward
		if isPortLessThan(pps[i].Port, pps[j].Port) {
			return true
		}
		// flip it around, run it the other way
		if isPortLessThan(pps[j].Port, pps[i].Port) {
			return false
		}
		// neither is less than the other?  fall back to protocol
		return pps[i].Protocol < pps[j].Protocol
	})
	return &SpecificPortMatcher{Ports: pps}
}

func (s *SpecificPortMatcher) Subtract(other *SpecificPortMatcher) (bool, *SpecificPortMatcher) {
	var remaining []*PortProtocolMatcher
	for _, thisPort := range s.Ports {
		found := false
		for _, otherPort := range other.Ports {
			if thisPort.Equals(otherPort) {
				found = true
				break
			}
		}
		if !found {
			remaining = append(remaining, thisPort)
		}
	}
	if len(remaining) == 0 {
		return true, nil
	}
	return false, &SpecificPortMatcher{Ports: remaining}
}

// isPortLessThan orders from low to high:
// - nil
// - string
// - int
func isPortLessThan(a *intstr.IntOrString, b *intstr.IntOrString) bool {
	if a == nil {
		return b != nil
	}
	if b == nil {
		return false
	}
	switch a.Type {
	case intstr.Int:
		switch b.Type {
		case intstr.Int:
			return a.IntVal < b.IntVal
		case intstr.String:
			return false
		default:
			panic("invalid type")
		}
	case intstr.String:
		switch b.Type {
		case intstr.Int:
			return true
		case intstr.String:
			return a.StrVal < b.StrVal
		default:
			panic("invalid type")
		}
	default:
		panic("invalid type")
	}
}

func isPortMatch(a intstr.IntOrString, portInt int, portName string) bool {
	switch a.Type {
	case intstr.Int:
		return int(a.IntVal) == portInt
	case intstr.String:
		return a.StrVal == portName
	default:
		panic("invalid type")
	}
}

func isIntStringEqual(a intstr.IntOrString, b intstr.IntOrString) bool {
	switch a.Type {
	case intstr.Int:
		switch b.Type {
		case intstr.Int:
			return a.IntVal == b.IntVal
		case intstr.String:
			return false
		default:
			panic("invalid type")
		}
	case intstr.String:
		switch b.Type {
		case intstr.Int:
			return false
		case intstr.String:
			return a.StrVal == b.StrVal
		default:
			panic("invalid type")
		}
	default:
		panic("invalid type")
	}
}
