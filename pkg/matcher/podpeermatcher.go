package matcher

import (
	v1 "k8s.io/api/core/v1"
)

type PodPeerMatcher struct {
	Namespace *NameLabelsSelector
	Pod       *NameLabelsSelector
	Port      *PortMatcher
}

func (ppm *PodPeerMatcher) PrimaryKey() string {
	return ppm.Namespace.GetPrimaryKey() + "---" + ppm.Pod.GetPrimaryKey()
}

func (ppm *PodPeerMatcher) Allows(peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool {
	if peer.IsExternal() {
		return false
	}
	return ppm.Namespace.Matches(peer.Internal.Namespace, peer.Internal.NamespaceLabels) &&
		ppm.Pod.Matches(peer.Internal.PodName, peer.Internal.PodLabels) &&
		ppm.Port.Allows(portInt, portName, protocol)
}

// PodMatcher possibilities:
// 1. PodSelector:
//   - empty/nil
//   - not empty
// 2. NamespaceSelector
//   - nil
//   - empty
//   - not empty
//
// Combined:
// 1. all pods in policy namespace
//   - empty/nil PodSelector
//   - nil NamespaceSelector
//
// 2. all pods in all namespaces
//   - empty/nil PodSelector
//   - empty NamespaceSelector
//
// 3. all pods in matching namespaces
//   - empty/nil PodSelector
//   - not empty NamespaceSelector
//
// 4. matching pods in policy namespace
//   - not empty PodSelector
//   - nil NamespaceSelector
//
// 5. matching pods in all namespaces
//   - not empty PodSelector
//   - empty NamespaceSelector
//
// 6. matching pods in matching namespaces
//   - not empty PodSelector
//   - not empty NamespaceSelector
//
// 7. everything
//   - don't have anything at all -- i.e. empty []NetworkPolicyPeer
//
