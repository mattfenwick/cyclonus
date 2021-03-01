package generator

import (
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

func (t *TestCaseGeneratorReplacement) ZeroPeersTestCases() []*TestCase {
	var cases []*TestCase
	for _, isIngress := range []bool{true, false} {
		dir := describeDirectionality(isIngress)
		cases = append(cases, NewSingleStepTestCase("", NewStringSet(dir, TagEmptyPeerSlice), ProbeAllAvailable,
			CreatePolicy(BuildPolicy(SetPeers(isIngress, emptySliceOfPeers)).NetworkPolicy())))
	}
	return cases
}

func describePeer(peer NetworkPolicyPeer) []string {
	if peer.IPBlock != nil {
		if len(peer.IPBlock.Except) == 0 {
			return []string{TagIPBlock}
		}
		return []string{TagIPBlockWithExcept}
	}
	return []string{
		describePeerNamespaceSelector(peer.NamespaceSelector),
		describePeerPodSelector(peer.PodSelector),
	}
}

func (t *TestCaseGeneratorReplacement) SinglePeersTestCases() []*TestCase {
	var cases []*TestCase
	for _, isIngress := range []bool{true, false} {
		for _, peer := range DefaultPeers(t.PodIP) {
			tags := append([]string{TagSinglePeerSlice, describeDirectionality(isIngress)}, describePeer(peer)...)
			cases = append(cases,
				NewSingleStepTestCase("", NewStringSet(tags...), ProbeAllAvailable,
					CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{peer})).NetworkPolicy())))
		}
	}
	return cases
}

func (t *TestCaseGeneratorReplacement) TwoPeersTestCases() []*TestCase {
	var cases []*TestCase
	for _, isIngress := range []bool{true, false} {
		for i, peer1 := range DefaultPeers(t.PodIP) {
			for j, peer2 := range DefaultPeers(t.PodIP) {
				if i < j {
					tags := append([]string{TagTwoPlusPeerSlice, describeDirectionality(isIngress)}, describePeer(peer1)...)
					tags = append(tags, describePeer(peer2)...)
					cases = append(cases,
						NewSingleStepTestCase("", NewStringSet(tags...), ProbeAllAvailable,
							CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{peer1, peer2})).NetworkPolicy())))
				}
			}
		}
	}
	return cases
}

func (t *TestCaseGeneratorReplacement) PeersTestCases() []*TestCase {
	return flatten(
		t.ZeroPeersTestCases(),
		t.SinglePeersTestCases(),
		t.TwoPeersTestCases())
}
