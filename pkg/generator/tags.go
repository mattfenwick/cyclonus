package generator

const (
	TagPathological = "pathological"

	TagCreatePolicy       = "create policy"
	TagDeletePolicy       = "delete policy"
	TagUpdatePolicy       = "update policy"
	TagCreatePod          = "create pod"
	TagDeletePod          = "delete pod"
	TagSetPodLabels       = "set pod labels"
	TagCreateNamespace    = "create namespace"
	TagDeleteNamespace    = "delete namespace"
	TagSetNamespaceLabels = "set namespace labels"

	TagExample = "example"

	TagUpstreamE2E = "upstream e2e"

	TagTargetNamespace   = "target namespace"
	TagTargetPodSelector = "target pod selector"

	TagIngress = "ingress"
	TagEgress  = "egress"

	TagDenyAll  = "deny all"
	TagAllowAll = "allow all"

	TagEmptyPeerSlice       = "empty peer slice"
	TagSinglePeerSlice      = "single peer slice"
	TagTwoPlusPeerSlice     = "two or more peer slice"
	TagAllPodsNilSelector   = "all pods nil selector"
	TagAllPodsEmptySelector = "all pods empty selector"
	TagPodsByLabel          = "pods by label"
	TagAllNamespaces        = "all namespaces"
	TagNamespacesByLabel    = "namespaces by label"
	TagPolicyNamespace      = "policy namespace"
	TagIPBlock              = "IP block"
	TagIPBlockWithExcept    = "IP block with except"

	TagEmptyPortSlice   = "empty port slice"
	TagSinglePortSlice  = "single port slice"
	TagTwoPlusPortSlice = "two or more port slice"
	TagNilPort          = "nil port"
	TagNumberedPort     = "numbered port"
	TagNamedPort        = "named port"
	TagNilProtocol      = "nil protocol"
	TagTCPProtocol      = "TCP protocol"
	TagUDPProtocol      = "UDP protocol"
	TagSCTPProtocol     = "SCTP protocol"
)

var AllTags = []string{
	TagPathological,

	TagExample,

	TagUpstreamE2E,

	TagTargetNamespace,
	TagTargetPodSelector,

	TagIngress,
	TagEgress,

	TagDenyAll,
	TagAllowAll,

	TagEmptyPeerSlice,
	TagSinglePeerSlice,
	TagTwoPlusPeerSlice,
	TagAllPodsNilSelector,
	TagAllPodsEmptySelector,
	TagPodsByLabel,
	TagAllNamespaces,
	TagNamespacesByLabel,
	TagPolicyNamespace,
	TagIPBlock,
	TagIPBlockWithExcept,

	TagEmptyPortSlice,
	TagSinglePortSlice,
	TagTwoPlusPortSlice,
	TagNilPort,
	TagNumberedPort,
	TagNamedPort,
	TagNilProtocol,
	TagTCPProtocol,
	TagUDPProtocol,
	TagSCTPProtocol,
}

type StringSet map[string]bool

func NewStringSet(elems ...string) StringSet {
	dict := map[string]bool{}
	for _, e := range elems {
		dict[e] = true
	}
	return dict
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
