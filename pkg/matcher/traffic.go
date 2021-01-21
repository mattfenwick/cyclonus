package matcher

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Traffic struct {
	Source      *TrafficPeer
	Destination *TrafficPeer

	PortProtocol *PortProtocol
}

type PortProtocol struct {
	Protocol v1.Protocol
	Port     intstr.IntOrString
}

type TrafficPeer struct {
	Internal *InternalPeer
	IP       string
}

func (p *TrafficPeer) Namespace() string {
	if p.Internal == nil {
		return ""
	}
	return p.Internal.Namespace
}

func (p *TrafficPeer) IsExternal() bool {
	return p.Internal == nil
}

type InternalPeer struct {
	PodLabels map[string]string
	//Pod             string
	NamespaceLabels map[string]string
	Namespace       string
	//NodeLabels      map[string]string
	//Node            string
}
