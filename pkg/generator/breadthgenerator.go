package generator

import (
	"fmt"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func baseBreadthPolicy() *Netpol {
	return &Netpol{
		Name: "base",
		Target: &NetpolTarget{
			Namespace:   "x",
			PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{"pod": "a"}},
		},
		Ingress: &NetpolPeers{Rules: []*Rule{{
			Ports: []NetworkPolicyPort{{
				Port:     &port80,
				Protocol: &tcp,
			}},
			Peers: []NetworkPolicyPeer{{
				PodSelector:       podBCMatchExpressionsSelector,
				NamespaceSelector: nsXYMatchExpressionsSelector},
			}},
		}},
		Egress: &NetpolPeers{Rules: []*Rule{
			{
				Ports: []NetworkPolicyPort{{
					Port:     &port80,
					Protocol: &tcp,
				}},
				Peers: []NetworkPolicyPeer{{
					PodSelector:       podABMatchExpressionsSelector,
					NamespaceSelector: nsYZMatchExpressionsSelector},
				},
			},
			AllowDNSRule,
		}},
	}
}

// BreadthGenerator should provide tests that cover the following features, without worrying about
//   corner cases or going into features in depth:
// - probe, policy on tcp
// - probe, policy on udp
// - probe, policy on sctp
// - named port
// - numbered port
// - pod selector (all, by label)
// - ingress (+ same for egress)
//   - deny all
//   - allow all
//   - pod
//     - namespace selector (all, by label, same as target)
//     - pod selector (all, by label)
//   - ipblock
//     - allow cidr
//     - except cidr
// - egress: DNS (udp/53)
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

func (e *BreadthGenerator) Policies() [][]Setter {
	var policies [][]Setter

	addPolicy := func(description string, setters ...Setter) {
		policies = append(policies, append([]Setter{SetDescription(description)}, setters...))
	}

	addPolicy("base policy")

	for _, isIngress := range []bool{true, false} {
		prefix := "ingress: "
		if !isIngress {
			prefix = "egress: "
		}

		addPolicy(prefix+"deny all", SetRules(isIngress, []*Rule{}))
		addPolicy(prefix+"allow all", SetRules(isIngress, []*Rule{{}}))

		addPolicy(prefix+"all pods and ip address", SetPeers(isIngress, emptySliceOfPeers))

		addPolicy(prefix+"all pods, policy namespace", SetPeers(isIngress, []NetworkPolicyPeer{{PodSelector: emptySelector, NamespaceSelector: nilSelector}}))
		addPolicy(prefix+"all pods, namespace by label", SetPeers(isIngress, []NetworkPolicyPeer{{PodSelector: emptySelector, NamespaceSelector: nsXMatchLabelsSelector}}))
		addPolicy(prefix+"all pods, all namespaces", SetPeers(isIngress, []NetworkPolicyPeer{{PodSelector: emptySelector, NamespaceSelector: emptySelector}}))
		addPolicy(prefix+"pods by label, policy namespace", SetPeers(isIngress, []NetworkPolicyPeer{{PodSelector: podCMatchLabelsSelector, NamespaceSelector: nilSelector}}))
		addPolicy(prefix+"pods by label, namespace by label", SetPeers(isIngress, []NetworkPolicyPeer{{PodSelector: podCMatchLabelsSelector, NamespaceSelector: nsXMatchLabelsSelector}}))
		addPolicy(prefix+"pods by label, all namespaces", SetPeers(isIngress, []NetworkPolicyPeer{{PodSelector: podCMatchLabelsSelector, NamespaceSelector: emptySelector}}))

		// TODO normalize these CIDRs
		cidr24 := fmt.Sprintf("%s/24", e.PodIP)
		cidr28 := fmt.Sprintf("%s/28", e.PodIP)
		addPolicy(prefix+"ipblock", SetPeers(isIngress, []NetworkPolicyPeer{{IPBlock: &IPBlock{CIDR: cidr24}}}))
		addPolicy(prefix+"ipblock with except", SetPeers(isIngress, []NetworkPolicyPeer{{IPBlock: &IPBlock{CIDR: cidr24, Except: []string{cidr28}}}}))
	}

	return policies
}

func (e *BreadthGenerator) ActionTestCases() []*TestCase {
	return []*TestCase{
		{
			Description: "Create/delete policy",
			Steps: []*TestStep{
				NewTestStep(ProbeAllAvailable, CreatePolicy(baseBreadthPolicy().NetworkPolicy())),
				NewTestStep(ProbeAllAvailable, DeletePolicy(baseBreadthPolicy().Target.Namespace, baseBreadthPolicy().Name)),
			},
		},
		{
			Description: "Create/update policy",
			Steps: []*TestStep{
				NewTestStep(ProbeAllAvailable, CreatePolicy(baseBreadthPolicy().NetworkPolicy())),
				NewTestStep(ProbeAllAvailable, UpdatePolicy(BuildPolicy(SetPorts(true, []NetworkPolicyPort{{Protocol: &udp, Port: &portServe81UDP}})).NetworkPolicy())),
				// TODO make an analogous modification for egress
			},
		},

		{
			Description: "Create/delete namespace",
			Steps: []*TestStep{
				NewTestStep(ProbeAllAvailable,
					CreatePolicy(baseBreadthPolicy().NetworkPolicy())),
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
					CreatePolicy(baseBreadthPolicy().NetworkPolicy())),
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
	var cases []*TestCase
	for _, modifications := range e.Policies() {
		policy := BuildPolicy(modifications...)
		cases = append(cases, NewSingleStepTestCase(policy.Description, NewStringSet(), ProbeAllAvailable, CreatePolicy(policy.NetworkPolicy())))
	}
	return append(cases, e.ActionTestCases()...)
}
