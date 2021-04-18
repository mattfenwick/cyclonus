package generator

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type peer struct {
	Description string
	Peer        NetworkPolicyPeer
}

func ipBlockPeers(podIP string) []*peer {
	cidrBut8 := kube.MakeCIDRFromZeroes(podIP, 8)
	cidrBut4 := kube.MakeCIDRFromZeroes(podIP, 4)
	return []*peer{
		{Description: "simple ipblock", Peer: NetworkPolicyPeer{IPBlock: &IPBlock{CIDR: cidrBut8}}},
		{Description: "ipblock with except", Peer: NetworkPolicyPeer{IPBlock: &IPBlock{CIDR: cidrBut8, Except: []string{cidrBut4}}}},
	}
}

func podPeers() []*peer {
	return []*peer{
		// skip this case -- this is where IPBlock needs to be non-nil
		//{Description: "single peer: ", Peer:        NetworkPolicyPeer{PodSelector:       nil, NamespaceSelector: nil}},
		{Description: "empty pods + nil ns", Peer: NetworkPolicyPeer{PodSelector: emptySelector, NamespaceSelector: nil}},
		{Description: "pods by label + nil ns", Peer: NetworkPolicyPeer{PodSelector: podCMatchLabelsSelector, NamespaceSelector: nil}},

		{Description: "nil pods + empty ns", Peer: NetworkPolicyPeer{PodSelector: nil, NamespaceSelector: emptySelector}},
		{Description: "empty pods + empty ns", Peer: NetworkPolicyPeer{PodSelector: emptySelector, NamespaceSelector: emptySelector}},
		{Description: "pods by label + empty ns", Peer: NetworkPolicyPeer{PodSelector: podCMatchLabelsSelector, NamespaceSelector: emptySelector}},

		{Description: "nil pods + ns by label", Peer: NetworkPolicyPeer{PodSelector: nil, NamespaceSelector: nsXMatchLabelsSelector}},
		{Description: "empty pods + ns by label", Peer: NetworkPolicyPeer{PodSelector: emptySelector, NamespaceSelector: nsXMatchLabelsSelector}},
		{Description: "pods by label + ns by label", Peer: NetworkPolicyPeer{PodSelector: podCMatchLabelsSelector, NamespaceSelector: nsXMatchLabelsSelector}},
	}
}

func makePeers(podIP string) []*peer {
	return append(podPeers(), ipBlockPeers(podIP)...)
}

func describePeerPodSelector(selector *metav1.LabelSelector) string {
	if selector == nil || kube.IsLabelSelectorEmpty(*selector) {
		return TagAllPods
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
		cases = append(cases, NewSingleStepTestCase(fmt.Sprintf("%s: empty peers", dir), NewStringSet(dir, TagAnyPeer), ProbeAllAvailable,
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
	peers := makePeers(t.PodIP)
	for _, isIngress := range []bool{true, false} {
		for _, p := range peers {
			tags := append(describePeer(p.Peer), describeDirectionality(isIngress))
			cases = append(cases,
				NewSingleStepTestCase(p.Description, NewStringSet(tags...), ProbeAllAvailable,
					CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{p.Peer})).NetworkPolicy())))
		}
	}
	return cases
}

func (t *TestCaseGenerator) TwoPeersTestCases() []*TestCase {
	var cases []*TestCase
	peers := makePeers(t.PodIP)
	for _, isIngress := range []bool{true, false} {
		for i, p1 := range peers {
			for j, p2 := range peers {
				if i < j {
					dir := describeDirectionality(isIngress)
					tags := append(describePeer(p1.Peer), TagMultiPeer, dir)
					tags = append(tags, describePeer(p2.Peer)...)
					cases = append(cases,
						NewSingleStepTestCase(fmt.Sprintf("%s, 2-peer: %s, %s", dir, p1.Description, p2.Description), NewStringSet(tags...), ProbeAllAvailable,
							CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{p1.Peer, p2.Peer})).NetworkPolicy())))
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
