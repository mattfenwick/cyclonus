package matcher

import (
	"encoding/json"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type PortMatcher interface {
	Allows(portInt int, portName string, protocol v1.Protocol) bool
}

func CombinePortMatchers(a PortMatcher, b PortMatcher) PortMatcher {
	switch l := a.(type) {
	case *AllPortMatcher:
		return a
	case *NonePortMatcher:
		return b
	case *SpecificPortMatcher:
		switch r := b.(type) {
		case *AllPortMatcher:
			return b
		case *NonePortMatcher:
			return a
		case *SpecificPortMatcher:
			return l.Combine(r)
		default:
			panic(errors.Errorf("invalid Port type %T", b))
		}
	default:
		panic(errors.Errorf("invalid Port type %T", a))
	}
}

type NonePortMatcher struct{}

func (n *NonePortMatcher) Allows(portInt int, portName string, protocol v1.Protocol) bool {
	return false
}

func (n *NonePortMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "no ports",
	})
}

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

func (p *PortProtocolMatcher) Allows(portInt int, portName string, protocol v1.Protocol) bool {
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
		if matcher.Allows(portInt, portName, protocol) {
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
	return &SpecificPortMatcher{Ports: pps}
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
