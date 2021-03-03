package generator

import (
	"github.com/pkg/errors"
	"sort"
	"strings"
)

const (
	TagAction        = "action"
	TagTarget        = "target"
	TagDirection     = "direction"
	TagPolicyStack   = "policy-stack"
	TagRule          = "rule"
	TagProtocol      = "protocol"
	TagPort          = "port"
	TagPeerIPBlock   = "peer-ipblock"
	TagPeerPods      = "peer-pods"
	TagMiscellaneous = "miscellaneous"
)

const (
	TagCreatePolicy       = "create-policy"
	TagDeletePolicy       = "delete-policy"
	TagUpdatePolicy       = "update-policy"
	TagCreatePod          = "create-pod"
	TagDeletePod          = "delete-pod"
	TagSetPodLabels       = "set-pod-labels"
	TagCreateNamespace    = "create-namespace"
	TagDeleteNamespace    = "delete-namespace"
	TagSetNamespaceLabels = "set-namespace-labels"
)

const (
	TagTargetNamespace   = "target-namespace"
	TagTargetPodSelector = "target-pod-selector"
)

const (
	TagIngress = "ingress"
	TagEgress  = "egress"
)

const (
	TagDenyAll           = "deny-all"
	TagAllowAll          = "allow-all"
	TagAnyPeer           = "any-peer"
	TagAnyPortProtocol   = "any-port-protocol"
	TagMultiPeer         = "multi-peer"
	TagMultiPortProtocol = "multi-port/protocol"
)

const (
	TagAllPods           = "all-pods"
	TagPodsByLabel       = "pods-by-label"
	TagAllNamespaces     = "all-namespaces"
	TagNamespacesByLabel = "namespaces-by-label"
	TagPolicyNamespace   = "policy-namespace"
)

const (
	TagIPBlockNoExcept   = "IP-block-no-except"
	TagIPBlockWithExcept = "IP-block-with-except"
)

const (
	TagAnyPort      = "any-port"
	TagNumberedPort = "numbered-port"
	TagNamedPort    = "named-port"
)

const (
	TagTCPProtocol  = "tcp"
	TagUDPProtocol  = "udp"
	TagSCTPProtocol = "sctp"
)

const (
	TagPathological = "pathological"
	TagConflict     = "conflict"
	TagExample      = "example"
	TagUpstreamE2E  = "upstream-e2e"
)

var AllTags = map[string][]string{
	TagAction: {
		TagCreatePolicy,
		TagDeletePolicy,
		TagUpdatePolicy,
		TagCreatePod,
		TagDeletePod,
		TagSetPodLabels,
		TagCreateNamespace,
		TagDeleteNamespace,
		TagSetNamespaceLabels,
	},
	TagTarget: {
		TagTargetNamespace,
		TagTargetPodSelector,
	},
	TagDirection: {
		TagIngress,
		TagEgress,
	},
	TagPolicyStack: {},
	TagRule: {
		TagDenyAll,
		TagAllowAll,
		TagAnyPeer,
		TagAnyPortProtocol,
		TagMultiPeer,
		TagMultiPortProtocol,
	},
	TagPeerPods: {
		TagAllPods,
		TagPodsByLabel,
		TagAllNamespaces,
		TagNamespacesByLabel,
		TagPolicyNamespace,
	},
	TagPeerIPBlock: {
		TagIPBlockNoExcept,
		TagIPBlockWithExcept,
	},
	TagPort: {
		TagAnyPort,
		TagNumberedPort,
		TagNamedPort,
	},
	TagProtocol: {
		TagTCPProtocol,
		TagUDPProtocol,
		TagSCTPProtocol,
	},
	TagMiscellaneous: {
		TagPathological,
		TagConflict,
		TagExample,
		TagUpstreamE2E,
	},
}

var TagSet = map[string]bool{}
var TagSlice []string
var TagSubToPrimary = map[string]string{}

func init() {
	for primary, subs := range AllTags {
		TagSet[primary] = true
		TagSlice = append(TagSlice, primary)
		for _, sub := range subs {
			TagSet[sub] = true
			TagSlice = append(TagSlice, sub)
			if prevPrimary, ok := TagSubToPrimary[sub]; ok {
				panic(errors.Errorf("subordinate tag %s has multiple owners: %s, %s", sub, prevPrimary, primary))
			}
			TagSubToPrimary[sub] = primary
		}
	}
	sort.Strings(TagSlice)
}

func CountTestCasesByTag(testCases []*TestCase) map[string]int {
	counts := map[string]int{}
	for tag := range TagSet {
		counts[tag] = 0
	}
	for _, tc := range testCases {
		for _, key := range tc.Tags.Keys() {
			counts[key]++
		}
	}
	return counts
}

func ValidateTags(tags []string) error {
	var invalid []string
	for _, tag := range tags {
		if _, ok := TagSet[tag]; !ok {
			invalid = append(invalid, tag)
		}
	}
	if len(invalid) > 0 {
		return errors.Errorf("invalid tags: %s", strings.Join(invalid, ", "))
	}
	return nil
}

func MustGetPrimaryTag(subordinateTag string) string {
	primary, ok := TagSubToPrimary[subordinateTag]
	if !ok {
		panic(errors.Errorf("no primary tag found for %s", subordinateTag))
	}
	return primary
}

type StringSet map[string]bool

func NewStringSet(elems ...string) StringSet {
	dict := StringSet(map[string]bool{})
	for _, e := range elems {
		dict.Add(e)
	}
	return dict
}

func (s StringSet) Add(key string) {
	s[key] = true
	s[MustGetPrimaryTag(key)] = true
}

func (s StringSet) GroupTags() map[string][]string {
	grouped := map[string][]string{}
	addGroup := func(key string) {
		if _, ok := grouped[key]; !ok {
			grouped[key] = []string{}
		}
	}
	for tag := range s {
		if _, ok := AllTags[tag]; ok {
			addGroup(tag)
		} else if primary, ok := TagSubToPrimary[tag]; ok {
			addGroup(primary)
			grouped[primary] = append(grouped[primary], tag)
		} else {
			panic(errors.Errorf("tag %s is neither primary nor subordinate", tag))
		}
	}
	return grouped
}

func (s StringSet) Keys() []string {
	var slice []string
	for k := range s {
		slice = append(slice, k)
	}
	return slice
}

func (s StringSet) ContainsAny(slice []string) bool {
	for _, e := range slice {
		if _, ok := s[e]; ok {
			return true
		}
	}
	return false
}
