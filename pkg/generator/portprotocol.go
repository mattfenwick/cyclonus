package generator

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type PortProtocol struct {
	Protocol v1.Protocol
	Port     intstr.IntOrString
}
