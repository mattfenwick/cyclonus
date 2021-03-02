package generator

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ipBlockPeers(podIP string) []NetworkPolicyPeer {
	// TODO normalize these cidrs
	cidr24 := fmt.Sprintf("%s/24", podIP)
	//cidr26 := fmt.Sprintf("%s/26", podIP)
	cidr28 := fmt.Sprintf("%s/28", podIP)
	//cidr30 := fmt.Sprintf("%s/30", podIP)
	return []NetworkPolicyPeer{
		{
			IPBlock: &IPBlock{
				CIDR:   cidr24,
				Except: nil,
			},
		},
		{
			IPBlock: &IPBlock{
				CIDR:   cidr24,
				Except: []string{cidr28},
			},
		},
	}
}

func podPeers() []NetworkPolicyPeer {
	var peers []NetworkPolicyPeer
	for _, nsSel := range []*metav1.LabelSelector{nil, emptySelector, nsXMatchLabelsSelector} {
		for _, podSel := range []*metav1.LabelSelector{nil, emptySelector, podCMatchLabelsSelector} {
			if nsSel == nil && podSel == nil {
				// skip this case -- this is where IPBlock needs to be non-nil
			} else {
				peers = append(peers, NetworkPolicyPeer{
					PodSelector:       podSel,
					NamespaceSelector: nsSel,
					IPBlock:           nil,
				})
			}
		}
	}
	return peers
}

func makePeers(podIP string) []NetworkPolicyPeer {
	return append(podPeers(), ipBlockPeers(podIP)...)
}

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

func (t *TestCaseGenerator) ZeroPeersTestCases() []*TestCase {
	var cases []*TestCase
	for _, isIngress := range []bool{true, false} {
		dir := describeDirectionality(isIngress)
		cases = append(cases, NewSingleStepTestCase("", NewStringSet(dir, TagEmptyPeerSlice), ProbeAllAvailable,
			CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{})).NetworkPolicy())))
	}
	return cases
}

func describePeer(peer NetworkPolicyPeer) []string {
	if peer.IPBlock != nil {
		if len(peer.IPBlock.Except) == 0 {
			return []string{TagIPBlockNoExcept}
		}
		return []string{TagIPBlockWithExcept}
	}
	return []string{
		describePeerNamespaceSelector(peer.NamespaceSelector),
		describePeerPodSelector(peer.PodSelector),
	}
}

func (t *TestCaseGenerator) SinglePeersTestCases() []*TestCase {
	var cases []*TestCase
	for _, isIngress := range []bool{true, false} {
		for _, peer := range makePeers(t.PodIP) {
			tags := append([]string{TagSinglePeerSlice, describeDirectionality(isIngress)}, describePeer(peer)...)
			cases = append(cases,
				NewSingleStepTestCase("", NewStringSet(tags...), ProbeAllAvailable,
					CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{peer})).NetworkPolicy())))
		}
	}
	return cases
}

func (t *TestCaseGenerator) TwoPeersTestCases() []*TestCase {
	var cases []*TestCase
	for _, isIngress := range []bool{true, false} {
		for i, peer1 := range makePeers(t.PodIP) {
			for j, peer2 := range makePeers(t.PodIP) {
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

func (t *TestCaseGenerator) PeersTestCases() []*TestCase {
	return flatten(
		t.ZeroPeersTestCases(),
		t.SinglePeersTestCases(),
		t.TwoPeersTestCases())
}
