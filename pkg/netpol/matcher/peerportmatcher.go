package matcher

import (
	"encoding/json"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PeerPortMatcher struct {
	Peer PeerMatcher
	Port PortMatcher
}

func (sdap *PeerPortMatcher) Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool {
	return sdap.Port.Allows(portProtocol) && sdap.Peer.Allows(peer)
}

// PeerPortMatcher possibilities:
// 1. PodSelector:
//   - empty/nil
//   - not empty
// 2. NamespaceSelector
//   - nil
//   - empty
//   - not empty
// 3. IPBlock
//   - nil
//   - not nil
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
// 7. matching IPBlock
//   - IPBlock
//
// 8. everything
//   - don't have anything at all -- i.e. empty []NetworkPolicyPeer
//

type PeerMatcher interface {
	Allows(peer *TrafficPeer) bool
}

// AllPodsInPolicyNamespacePeerMatcher models the case where in NetworkPolicyPeer:
// - PodSelector is empty or nil
// - NamespaceSelector is nil
// - IPBlock is nil
type AllPodsInPolicyNamespacePeerMatcher struct {
	Namespace string
}

func (p *AllPodsInPolicyNamespacePeerMatcher) Allows(peer *TrafficPeer) bool {
	if peer.IsExternal() {
		return false
	}
	return peer.Namespace() == p.Namespace
}

func (p *AllPodsInPolicyNamespacePeerMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":      "all pods in policy namespace",
		"Namespace": p.Namespace,
	})
}

// AllPodsAllNamespacesPeerMatcher models the case where in NetworkPolicyPeer:
// - PodSelector is nil or empty
// - NamespaceSelector is empty (but not nil!)
// - IPBlock is nil
type AllPodsAllNamespacesPeerMatcher struct{}

func (a *AllPodsAllNamespacesPeerMatcher) Allows(peer *TrafficPeer) bool {
	return !peer.IsExternal()
}

func (a *AllPodsAllNamespacesPeerMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "all pods in all namespaces",
	})
}

// AllPodsInMatchingNamespacesPeerMatcher models the case where in NetworkPolicyPeer:
// - PodSelector is nil or empty
// - NamespaceSelector is not empty
// - IPBlock is nil
type AllPodsInMatchingNamespacesPeerMatcher struct {
	NamespaceSelector metav1.LabelSelector
}

func (a *AllPodsInMatchingNamespacesPeerMatcher) Allows(peer *TrafficPeer) bool {
	if peer.IsExternal() {
		return false
	}
	return kube.IsLabelsMatchLabelSelector(peer.Internal.NamespaceLabels, a.NamespaceSelector)
}

func (a *AllPodsInMatchingNamespacesPeerMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":              "all pods in matching namespaces",
		"NamespaceSelector": a.NamespaceSelector,
	})
}

// MatchingPodsInPolicyNamespacePeerMatcher models the case where in NetworkPolicyPeer:
// - PodSelector is not empty
// - NamespaceSelector is nil
// - IPBlock is nil
type MatchingPodsInPolicyNamespacePeerMatcher struct {
	PodSelector metav1.LabelSelector
	Namespace   string
}

func (p *MatchingPodsInPolicyNamespacePeerMatcher) Allows(peer *TrafficPeer) bool {
	if peer.IsExternal() {
		return false
	}
	return kube.IsLabelsMatchLabelSelector(peer.Internal.PodLabels, p.PodSelector) && peer.Namespace() == p.Namespace
}

func (p *MatchingPodsInPolicyNamespacePeerMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":        "matching pods in policy namespace",
		"PodSelector": p.PodSelector,
		"Namespace":   p.Namespace,
	})
}

// MatchingPodsInAllNamespacesPeerMatcher models the case where in NetworkPolicyPeer:
// - PodSelector is not nil
// - NamespaceSelector is empty (but not nil!)
// - IPBlock is nil
type MatchingPodsInAllNamespacesPeerMatcher struct {
	PodSelector metav1.LabelSelector
}

func (p *MatchingPodsInAllNamespacesPeerMatcher) Allows(peer *TrafficPeer) bool {
	if peer.IsExternal() {
		return false
	}
	return kube.IsLabelsMatchLabelSelector(peer.Internal.PodLabels, p.PodSelector)
}

func (p *MatchingPodsInAllNamespacesPeerMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":        "pods in all namespaces",
		"PodSelector": p.PodSelector,
	})
}

// MatchingPodsInMatchingNamespacesPeerMatcher models the case where in NetworkPolicyPeer:
// - PodSelector is not nil
// - NamespaceSelector is not empty
// - IPBlock is nil
type MatchingPodsInMatchingNamespacesPeerMatcher struct {
	PodSelector       metav1.LabelSelector
	NamespaceSelector metav1.LabelSelector
}

func (s *MatchingPodsInMatchingNamespacesPeerMatcher) Allows(peer *TrafficPeer) bool {
	if peer.IsExternal() {
		return false
	}
	internal := peer.Internal
	return kube.IsLabelsMatchLabelSelector(internal.NamespaceLabels, s.NamespaceSelector) &&
		kube.IsLabelsMatchLabelSelector(internal.PodLabels, s.PodSelector)
}

func (s *MatchingPodsInMatchingNamespacesPeerMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":              "matching pods in matching namespaces",
		"PodSelector":       s.PodSelector,
		"NamespaceSelector": s.NamespaceSelector,
	})
}

// AnywherePeerMatcher models the case where NetworkPolicy(E|In)gressRule.(From|To) is empty
type AnywherePeerMatcher struct{}

func (a *AnywherePeerMatcher) Allows(peer *TrafficPeer) bool {
	return true
}

func (a *AnywherePeerMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "anywhere",
	})
}

// IPBlockPeerMatcher models the case where IPBlock is not nil, and both
// PodSelector and NamespaceSelector are nil
type IPBlockPeerMatcher struct {
	IPBlock *networkingv1.IPBlock
}

func (ibsd *IPBlockPeerMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":   "IPBlock",
		"CIDR":   ibsd.IPBlock.CIDR,
		"Except": ibsd.IPBlock.Except,
	})
}

func (ibsd *IPBlockPeerMatcher) Allows(peer *TrafficPeer) bool {
	return kube.IsIPBlockMatchForIP(peer.IP, ibsd.IPBlock)
}
