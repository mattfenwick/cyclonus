package generator

import (
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BreadthGenerator struct {
	PodIP    string
	AllowDNS bool
}

func NewBreadthGenerator(allowDNS bool, podIP string) *BreadthGenerator {
	return &BreadthGenerator{
		PodIP:    podIP,
		AllowDNS: allowDNS,
	}
}

func (e *BreadthGenerator) ActionTestCases() []*TestCase {
	return []*TestCase{
		{
			Description: "Create/delete policy",
			Steps: []*TestStep{
				NewTestStep(ProbeAllAvailable, CreatePolicy(baseTestPolicy().NetworkPolicy())),
				NewTestStep(ProbeAllAvailable, DeletePolicy(baseTestPolicy().Target.Namespace, baseTestPolicy().Name)),
			},
		},
		{
			Description: "Create/update policy",
			Steps: []*TestStep{
				NewTestStep(ProbeAllAvailable, CreatePolicy(baseTestPolicy().NetworkPolicy())),
				NewTestStep(ProbeAllAvailable, UpdatePolicy(BuildPolicy(SetPorts(true, []NetworkPolicyPort{{Protocol: &udp, Port: &portServe81UDP}})).NetworkPolicy())),
				// TODO make an analogous modification for egress
			},
		},

		{
			Description: "Create/delete namespace",
			Steps: []*TestStep{
				NewTestStep(ProbeAllAvailable,
					CreatePolicy(baseTestPolicy().NetworkPolicy())),
				NewTestStep(ProbeAllAvailable,
					CreateNamespace("y-2", map[string]string{"ns": "y"}),
					CreatePod("y-2", "a", map[string]string{"pod": "a"}),
					CreatePod("y-2", "b", map[string]string{"pod": "b"})),
				NewTestStep(ProbeAllAvailable, DeleteNamespace("y-2")),
			},
		},
		{
			Description: "Update namespace so that policy applies, then again so it no longer applies",
			Steps: []*TestStep{
				NewTestStep(ProbeAllAvailable,
					CreatePolicy(BuildPolicy(SetPeers(true, []NetworkPolicyPeer{{
						NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"new-ns": "qrs"}}}})).NetworkPolicy())),
				NewTestStep(ProbeAllAvailable,
					SetNamespaceLabels("y", map[string]string{"ns": "y", "new-ns": "qrs"})),
				NewTestStep(ProbeAllAvailable,
					SetNamespaceLabels("y", map[string]string{"ns": "y"})),
			},
		},

		{
			Description: "Create/delete pod",
			Steps: []*TestStep{
				NewTestStep(ProbeAllAvailable,
					CreatePolicy(baseTestPolicy().NetworkPolicy())),
				NewTestStep(ProbeAllAvailable,
					CreatePod("x", "d", map[string]string{"pod": "d"})),
				NewTestStep(ProbeAllAvailable,
					DeletePod("x", "d")),
			},
		},
		{
			Description: "Update pod so that policy applies, then again so it no longer applies",
			Steps: []*TestStep{
				NewTestStep(ProbeAllAvailable,
					CreatePolicy(BuildPolicy(SetPeers(true, []NetworkPolicyPeer{{
						PodSelector:       &metav1.LabelSelector{MatchLabels: map[string]string{"new-label": "abc"}},
						NamespaceSelector: nsYZMatchExpressionsSelector}})).NetworkPolicy())),
				NewTestStep(ProbeAllAvailable,
					SetPodLabels("y", "b", map[string]string{"pod": "b", "new-label": "abc"})),
				NewTestStep(ProbeAllAvailable,
					SetPodLabels("y", "b", map[string]string{"pod": "b"})),
			},
		},
	}
}

func (e *BreadthGenerator) GenerateTestCases() []*TestCase {
	return e.ActionTestCases()
}
