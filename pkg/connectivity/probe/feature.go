package probe

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

// Goal: track what's used in probes.  Problem: AllAvailable could include SCTP/UDP/TCP etc., which
//    depend on the resources and which this package doesn't know about.
const (
	//FeatureAllAvailable  = "AllAvailable"
	//FeatureNumberedPort = "probe on numbered port"
	//FeatureNamedPort    = "probe on named port"
	FeatureTCP  = "probe on TCP"
	FeatureUDP  = "probe on UDP"
	FeatureSCTP = "probe on SCTP"
)

func ProtocolToFeature(protocol v1.Protocol) string {
	switch protocol {
	case v1.ProtocolSCTP:
		return FeatureSCTP
	case v1.ProtocolUDP:
		return FeatureUDP
	case v1.ProtocolTCP:
		return FeatureTCP
	default:
		panic(errors.Errorf("invalid protocol %+v", protocol))
	}
}
