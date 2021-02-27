package generator

import (
	"fmt"
	. "k8s.io/api/networking/v1"
)

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

			NewSingleStepTestCase("", NewStringSet(dir, TagAllPods, TagPolicyNamespace), ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{{PodSelector: emptySelector, NamespaceSelector: nilSelector}})).NetworkPolicy())),
			NewSingleStepTestCase("", NewStringSet(dir, TagAllPods, TagNamespacesByLabel), ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{{PodSelector: emptySelector, NamespaceSelector: nsXMatchLabelsSelector}})).NetworkPolicy())),
			NewSingleStepTestCase("", NewStringSet(dir, TagAllPods, TagAllNamespaces), ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{{PodSelector: emptySelector, NamespaceSelector: emptySelector}})).NetworkPolicy())),

			NewSingleStepTestCase("", NewStringSet(dir, TagPodsByLabel, TagPolicyNamespace), ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{{PodSelector: podCMatchLabelsSelector, NamespaceSelector: nilSelector}})).NetworkPolicy())),
			NewSingleStepTestCase("", NewStringSet(dir, TagPodsByLabel, TagNamespacesByLabel), ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{{PodSelector: podCMatchLabelsSelector, NamespaceSelector: nsXMatchLabelsSelector}})).NetworkPolicy())),
			NewSingleStepTestCase("", NewStringSet(dir, TagPodsByLabel, TagAllNamespaces), ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{{PodSelector: podCMatchLabelsSelector, NamespaceSelector: emptySelector}})).NetworkPolicy())),

			NewSingleStepTestCase("", NewStringSet(dir, TagIPBlock), ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{{IPBlock: &IPBlock{CIDR: cidr24}}})).NetworkPolicy())),
			NewSingleStepTestCase("", NewStringSet(dir, TagIPBlockWithExcept), ProbeAllAvailable,
				CreatePolicy(BuildPolicy(SetPeers(isIngress, []NetworkPolicyPeer{{IPBlock: &IPBlock{CIDR: cidr24, Except: []string{cidr28}}}})).NetworkPolicy())))
	}
	return cases
}
