package generator

import (
	networkingv1 "k8s.io/api/networking/v1"
)

func (t *TestCaseGenerator) NamespaceTestCases() []*TestCase {
	peers := []networkingv1.NetworkPolicyPeer{{NamespaceSelector: nsZMatchDefaultLabelsSelector}}
	return []*TestCase{
		NewSingleStepTestCase("ingress: select namespace by default label",
			NewStringSet(TagNamespacesByDefaultLabel, TagIngress),
			ProbeAllAvailable,
			CreatePolicy(BuildPolicy(SetPeers(true, peers)).NetworkPolicy())),
		NewSingleStepTestCase("egress: select namespace by default label",
			NewStringSet(TagNamespacesByDefaultLabel, TagEgress),
			ProbeAllAvailable,
			CreatePolicy(BuildPolicy(SetPeers(false, peers)).NetworkPolicy())),
	}
}
