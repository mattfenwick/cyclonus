package kube

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

func ParseProtocol(protocol string) (v1.Protocol, error) {
	switch protocol {
	case "tcp", "TCP":
		return v1.ProtocolTCP, nil
	case "udp", "UDP":
		return v1.ProtocolUDP, nil
	case "sctp", "SCTP":
		return v1.ProtocolSCTP, nil
	default:
		return "", errors.Errorf("invalid protocol %s", protocol)
	}
}
