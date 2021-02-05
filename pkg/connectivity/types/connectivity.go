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

type Answer struct {
	Ingress Connectivity
	Egress  Connectivity
}

func (a *Answer) ShortString() string {
	return a.Combined().ShortString()
}

func (a *Answer) Combined() Connectivity {
	switch a.Egress {
	case ConnectivityBlocked:
		return ConnectivityBlocked
	case ConnectivityAllowed:
		return a.Ingress
	case ConnectivityUnknown:
		switch a.Ingress {
		case ConnectivityAllowed:
			return ConnectivityUnknown
		default:
			return a.Ingress
		}
	default:
		panic(errors.Errorf("invalid Egress value %+v", a.Egress))
	}
}
