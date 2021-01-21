package matcher

import (
	"encoding/json"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type PortMatcher interface {
	Allows(port intstr.IntOrString, protocol v1.Protocol) bool
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
			return &SpecificPortMatcher{Ports: append(l.Ports, r.Ports...)}
		default:
			panic(errors.Errorf("invalid Port type %T", b))
		}
	default:
		panic(errors.Errorf("invalid Port type %T", a))
	}
}

type NonePortMatcher struct{}

func (n *NonePortMatcher) Allows(port intstr.IntOrString, protocol v1.Protocol) bool {
	return false
}

func (n *NonePortMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "no ports",
	})
}

type AllPortMatcher struct{}

func (ap *AllPortMatcher) Allows(port intstr.IntOrString, protocol v1.Protocol) bool {
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

func (ppm *PortProtocolMatcher) Allows(port intstr.IntOrString, protocol v1.Protocol) bool {
	if ppm.Port != nil {
		return isPortMatch(*ppm.Port, port) && ppm.Protocol == protocol
	}
	return ppm.Protocol == protocol
}

// SpecificPortMatcher models the case where traffic must match a named or numbered port
type SpecificPortMatcher struct {
	Ports []*PortProtocolMatcher
}

func (epp *SpecificPortMatcher) Allows(port intstr.IntOrString, protocol v1.Protocol) bool {
	for _, matcher := range epp.Ports {
		if matcher.Allows(port, protocol) {
			return true
		}
	}
	return false
}

func (epp *SpecificPortMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":  "specific ports",
		"Ports": epp.Ports,
	})
}

func isPortMatch(a intstr.IntOrString, b intstr.IntOrString) bool {
	switch a.Type {
	case intstr.Int:
		switch b.Type {
		case intstr.Int:
			return a.IntVal == b.IntVal
		case intstr.String:
			// TODO what if this named port resolves to same int?
			return false
		default:
			panic("invalid type")
		}
	case intstr.String:
		switch b.Type {
		case intstr.Int:
			// TODO what if this named port resolves to same int?
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
