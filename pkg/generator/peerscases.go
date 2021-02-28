package generator

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func describePeerPodSelector(selector *metav1.LabelSelector) string {
	if selector == nil {
		return TagAllPodsNilSelector
	} else if kube.IsLabelSelectorEmpty(*selector) {
		return TagAllPodsEmptySelector
	} else {
		return TagPodsByLabel
	}
}

func describePeerNamespaceSelector(selector *metav1.LabelSelector) string {
	if selector == nil {
		return TagPolicyNamespace
	} else if kube.IsLabelSelectorEmpty(*selector) {
		return TagAllNamespaces
	} else {
		return TagNamespacesByLabel
	}
}

func (t *TestCaseGeneratorReplacement) PeersTestCases() []*TestCase {
	var cases []*TestCase

	// TODO normalize these CIDRs
	cidr24 := fmt.Sprintf("%s/24", t.PodIP)
	cidr28 := fmt.Sprintf("%s/28", t.PodIP)

	for _, isIngress := range []bool{true, false} {
		dir := describeDirectionality(isIngress)
		cases = append(cases,
			NewSingleStepTestCase("", NewStringSet(dir, TagEmptyPeerSlice), ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPeers(isIngress, emptySliceOfPeers)).NetworkPolicy())),

			NewSingleStepTestCase("", NewStringSet(dir, TagIPBlock), ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{{IPBlock: &IPBlock{CIDR: cidr24}}})).NetworkPolicy())),
			NewSingleStepTestCase("", NewStringSet(dir, TagIPBlockWithExcept), ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{{IPBlock: &IPBlock{CIDR: cidr24, Except: []string{cidr28}}}})).NetworkPolicy())))

		for _, peers := range DefaultPodPeers() {
			cases = append(cases, NewSingleStepTestCase("",
				NewStringSet(dir, describePeerNamespaceSelector(peers.NamespaceSelector), describePeerPodSelector(peers.PodSelector)),
				ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{peers})).NetworkPolicy())))
		}
	}
	return cases
}
