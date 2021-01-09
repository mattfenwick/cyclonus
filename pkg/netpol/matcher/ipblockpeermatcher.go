package matcher

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	networkingv1 "k8s.io/api/networking/v1"
	"sort"
	"strings"
)

// IPBlockPeerMatcher models the case where IPBlock is not nil, and both
// PodSelector and NamespaceSelector are nil
type IPBlockPeerMatcher struct {
	IPBlock     *networkingv1.IPBlock
	PortMatcher PortMatcher
}

// PrimaryKey returns a content-based, deterministic key based on the IPBlock's
// CIDR and excepts.
func (ibsd *IPBlockPeerMatcher) PrimaryKey() string {
	block := ibsd.IPBlock
	var except []string
	for _, e := range block.Except {
		except = append(except, e)
	}
	sort.Slice(except, func(i, j int) bool {
		return except[i] < except[j]
	})
	return fmt.Sprintf("%s: [%s]", block.CIDR, strings.Join(except, ", "))
}

func (ibsd *IPBlockPeerMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":   "IPBlock",
		"CIDR":   ibsd.IPBlock.CIDR,
		"Except": ibsd.IPBlock.Except,
		"Port":   ibsd.PortMatcher,
	})
}

func (ibsd *IPBlockPeerMatcher) Allows(ip string, portProtocol *PortProtocol) bool {
	return kube.IsIPBlockMatchForIP(ip, ibsd.IPBlock) &&
		ibsd.PortMatcher.Allows(portProtocol.Port, portProtocol.Protocol)
}

func (ibsd *IPBlockPeerMatcher) Combine(other *IPBlockPeerMatcher) *IPBlockPeerMatcher {
	if ibsd.PrimaryKey() != other.PrimaryKey() {
		panic(errors.Errorf("unable to combine IPBlockPeerMatcher values with different primary keys: %s vs %s", ibsd.PrimaryKey(), other.PrimaryKey()))
	}
	return &IPBlockPeerMatcher{
		IPBlock:     ibsd.IPBlock,
		PortMatcher: CombinePortMatchers(ibsd.PortMatcher, other.PortMatcher),
	}
}
