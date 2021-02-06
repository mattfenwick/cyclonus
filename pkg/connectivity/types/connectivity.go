package types

import "github.com/pkg/errors"

type Connectivity string

const (
	ConnectivityUnknown             Connectivity = "unknown"
	ConnectivityCheckFailed         Connectivity = "checkfailed"
	ConnectivityInvalidNamedPort    Connectivity = "invalidnamedport"
	ConnectivityInvalidPortProtocol Connectivity = "invalidportprotocol"
	ConnectivityBlocked             Connectivity = "blocked"
	ConnectivityAllowed             Connectivity = "allowed"
)

var AllConnectivity = []Connectivity{
	ConnectivityUnknown,
	ConnectivityCheckFailed,
	ConnectivityInvalidNamedPort,
	ConnectivityInvalidPortProtocol,
	ConnectivityBlocked,
	ConnectivityAllowed,
}

func (p Connectivity) ShortString() string {
	switch p {
	case ConnectivityUnknown:
		return "?"
	case ConnectivityCheckFailed:
		return "!"
	case ConnectivityBlocked:
		return "X"
	case ConnectivityAllowed:
		return "."
	case ConnectivityInvalidNamedPort:
		return "P"
	case ConnectivityInvalidPortProtocol:
		return "N"
	default:
		panic(errors.Errorf("invalid Connectivity value: %+v", p))
	}
}

func CombineIngressEgressConnectivity(ingress Connectivity, egress Connectivity) Connectivity {
	switch egress {
	case ConnectivityBlocked:
		return ConnectivityBlocked
	case ConnectivityAllowed:
		return ingress
	case ConnectivityUnknown:
		switch ingress {
		case ConnectivityAllowed:
			return ConnectivityUnknown
		default:
			return ingress
		}
	default:
		panic(errors.Errorf("invalid egress Connectivity value %+v", egress))
	}
}
