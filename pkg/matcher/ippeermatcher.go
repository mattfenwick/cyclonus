package matcher

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"strings"
)

// IPPeerMatcher models the case where IPBlock is not nil, and both
// PodSelector and NamespaceSelector are nil
type IPPeerMatcher struct {
	IPBlock *networkingv1.IPBlock
	Port    PortMatcher
}

// PrimaryKey returns a content-based, deterministic key based on the IPBlock's
// CIDR and excepts.
func (i *IPPeerMatcher) PrimaryKey() string {
	block := i.IPBlock
	except := slice.Sort(i.IPBlock.Except)
	return fmt.Sprintf("%s: [%s]", block.CIDR, strings.Join(except, ", "))
}

func (i *IPPeerMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":   "IPBlock",
		"CIDR":   i.IPBlock.CIDR,
		"Except": i.IPBlock.Except,
		"Port":   i.Port,
	})
}

func (i *IPPeerMatcher) Allows(peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool {
	isIpMatch, err := kube.IsIPAddressMatchForIPBlock(peer.IP, i.IPBlock)
	// TODO propagate this error instead of panic
	if err != nil {
		panic(err)
	}
	return isIpMatch && i.Port.Allows(portInt, portName, protocol)
}
