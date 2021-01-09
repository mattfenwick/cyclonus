package matcher

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodPeerMatcher struct {
	Namespace NamespaceMatcher
	Pod       PodMatcher
	Port      PortMatcher
}

func (ppm *PodPeerMatcher) PrimaryKey() string {
	return ppm.Namespace.PrimaryKey() + "---" + ppm.Pod.PrimaryKey()
}

func (ppm *PodPeerMatcher) Allows(peer *InternalPeer, portProtocol *PortProtocol) bool {
	return ppm.Namespace.Allows(peer.Namespace, peer.NamespaceLabels) &&
		ppm.Pod.Allows(peer.PodLabels) &&
		ppm.Port.Allows(portProtocol.Port, portProtocol.Protocol)
}

func (ppm *PodPeerMatcher) Combine(otherPort PortMatcher) *PodPeerMatcher {
	return &PodPeerMatcher{
		Namespace: ppm.Namespace,
		Pod:       ppm.Pod,
		Port:      CombinePortMatchers(ppm.Port, otherPort),
	}
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

type PodMatcher interface {
	Allows(podLabels map[string]string) bool
	PrimaryKey() string
}

type AllPodMatcher struct{}

func (p *AllPodMatcher) Allows(podLabels map[string]string) bool {
	return true
}

func (p *AllPodMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "all pods",
	})
}

func (p *AllPodMatcher) PrimaryKey() string {
	return `{"type": "all-pods"}`
}

type LabelSelectorPodMatcher struct {
	Selector metav1.LabelSelector
}

func (p *LabelSelectorPodMatcher) Allows(podLabels map[string]string) bool {
	return kube.IsLabelsMatchLabelSelector(podLabels, p.Selector)
}

func (p *LabelSelectorPodMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":     "matching pods by label",
		"Selector": p.Selector,
	})
}

func (p *LabelSelectorPodMatcher) PrimaryKey() string {
	return fmt.Sprintf(`{"type": "label-selector", "selector": "%s"}`, SerializeLabelSelector(p.Selector))
}

// namespaces

type NamespaceMatcher interface {
	Allows(namespace string, namespaceLabels map[string]string) bool
	PrimaryKey() string
}

type ExactNamespaceMatcher struct {
	Namespace string
}

func (p *ExactNamespaceMatcher) Allows(namespace string, namespaceLabels map[string]string) bool {
	return p.Namespace == namespace
}

func (p *ExactNamespaceMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":      "specific namespace",
		"Namespace": p.Namespace,
	})
}

func (p *ExactNamespaceMatcher) PrimaryKey() string {
	return fmt.Sprintf(`{"type": "exact-namespace", "namespace": "%s"}`, p.Namespace)
}

type LabelSelectorNamespaceMatcher struct {
	Selector metav1.LabelSelector
}

func (p *LabelSelectorNamespaceMatcher) Allows(namespace string, namespaceLabels map[string]string) bool {
	return kube.IsLabelsMatchLabelSelector(namespaceLabels, p.Selector)
}

func (p *LabelSelectorNamespaceMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":     "matching namespace by label",
		"Selector": p.Selector,
	})
}

func (p *LabelSelectorNamespaceMatcher) PrimaryKey() string {
	return fmt.Sprintf(`{"type": "label-selector", "selector": "%s"}`, SerializeLabelSelector(p.Selector))
}

type AllNamespaceMatcher struct{}

func (a *AllNamespaceMatcher) Allows(namespace string, namespaceLabels map[string]string) bool {
	return true
}

func (a *AllNamespaceMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "all namespaces",
	})
}

func (a *AllNamespaceMatcher) PrimaryKey() string {
	return `{"type": "all-namespaces"}`
}
