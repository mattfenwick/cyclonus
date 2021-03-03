package generator

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
Conflicts:
 - for traffic: allow ingress side, deny egress side (or vice versa)
 - for ingress (or egress):
   - deny all, plus allow layered on top
   - deny all ips, allow all pods
   - deny all pods, allow all ips
     - allow all ips with 0.0.0.0/0
     - is there another way to allow all ips?
   - allow CIDR, deny smaller CIDR

Denies: allow something else, and not allowing the specific traffic, should deny traffic
 - allow nothing
 - allow different port
 - allow different protocol
 - allow different IP
 - allow different namespace/pod
*/

// allow all implicit (i.e. no rules)

var (
	ExplicitAllowAll = &NetpolPeers{
		Rules: []*Rule{
			{},
		},
	}
	DenyAll = &NetpolPeers{
		Rules: nil,
	}
	// DenyAll2 should be identical to DenyAll -- but just in case :)
	DenyAll2 = &NetpolPeers{
		Rules: []*Rule{},
	}

	AllowAllPodsRule = &Rule{
		Peers: []networkingv1.NetworkPolicyPeer{
			{
				NamespaceSelector: &metav1.LabelSelector{},
			},
		},
	}

	AllowAllByPod = &NetpolPeers{
		Rules: []*Rule{AllowAllPodsRule},
	}

	AllowAllByIPRule = &Rule{
		Peers: []networkingv1.NetworkPolicyPeer{
			{
				IPBlock: &networkingv1.IPBlock{
					CIDR: "0.0.0.0/0",
				},
			},
		},
	}

	AllowAllByIP = &NetpolPeers{
		Rules: []*Rule{AllowAllByIPRule},
	}

	DenyAllByIPRule = &Rule{
		Peers: []networkingv1.NetworkPolicyPeer{{
			IPBlock: &networkingv1.IPBlock{CIDR: "0.0.0.0/31"},
		}},
	}

	DenyAllByIP = &NetpolPeers{
		Rules: []*Rule{DenyAllByIPRule},
	}

	DenyAllByPodRule = &Rule{
		Peers: []networkingv1.NetworkPolicyPeer{{
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"this-will-never-happen": "qrs123"},
			},
		}},
	}

	DenyAllByPod = &NetpolPeers{
		Rules: []*Rule{DenyAllByPodRule},
	}
)

type conflictCase struct {
	Description string
	Tags        []string
	Policies    []*Netpol
}

func AllowAllIngressDenyAllEgress(source *NetpolTarget, dest *NetpolTarget) *conflictCase {
	return &conflictCase{
		Description: "deny all from source, allow all to dest",
		Tags:        []string{TagDenyAll, TagAllowAll, TagIngress, TagEgress},
		Policies: []*Netpol{
			{Name: "deny-all-egress", Target: source, Egress: DenyAll},
			{Name: "allow-all-ingress", Target: dest, Ingress: ExplicitAllowAll}}}
}

func AllowAllEgressDenyAllIngress(source *NetpolTarget, dest *NetpolTarget) *conflictCase {
	return &conflictCase{
		Description: "allow all from source, deny all to dest",
		Tags:        []string{TagDenyAll, TagAllowAll, TagIngress, TagEgress},
		Policies: []*Netpol{
			{Name: "allow-all-egress", Target: source, Egress: ExplicitAllowAll},
			{Name: "deny-all-ingress", Target: dest, Ingress: DenyAll},
		}}
}

func DenyAllEgressAllowAllEgress(source *NetpolTarget) *conflictCase {
	return &conflictCase{
		Description: "deny all + allow all from same source",
		Tags:        []string{TagDenyAll, TagAllowAll, TagEgress},
		Policies: []*Netpol{
			{Name: "deny-all-egress", Target: source, Egress: DenyAll},
			{Name: "allow-all-egress", Target: source, Egress: ExplicitAllowAll},
		}}
}

func DenyAllIngressAllowAllIngress(dest *NetpolTarget) *conflictCase {
	return &conflictCase{
		Description: "deny all + allow all to same dest",
		Tags:        []string{TagDenyAll, TagAllowAll, TagIngress},
		Policies: []*Netpol{
			{Name: "deny-all-ingress", Target: dest, Ingress: DenyAll},
			{Name: "allow-all-ingress", Target: dest, Ingress: ExplicitAllowAll},
		}}
}

func DenyAllEgressAllowAllEgressByPod(source *NetpolTarget) *conflictCase {
	return &conflictCase{
		Description: "deny all + allow all by pod from same source",
		Tags:        []string{TagDenyAll, TagAllPods, TagAllNamespaces, TagEgress},
		Policies: []*Netpol{
			{Name: "deny-all-egress", Target: source, Egress: DenyAll},
			{Name: "allow-all-egress-by-pod", Target: source, Egress: AllowAllByPod},
		}}
}

func DenyAllEgressAllowAllEgressByIP(source *NetpolTarget) *conflictCase {
	return &conflictCase{
		Description: "deny all + allow all by IP from same source",
		Tags:        []string{TagDenyAll, TagEgress},
		Policies: []*Netpol{
			{Name: "deny-all-egress", Target: source, Egress: DenyAll},
			{Name: "allow-all-egress-by-ip", Target: source, Egress: AllowAllByIP},
		}}
}

func DenyAllEgressByIPAllowAllEgressByPod(source *NetpolTarget) *conflictCase {
	return &conflictCase{
		Description: "deny all by IP + allow all by pod from same source",
		Tags:        []string{TagAllPods, TagAllNamespaces, TagEgress},
		Policies: []*Netpol{
			{Name: "deny-all-egress-by-ip", Target: source, Egress: DenyAllByIP},
			{Name: "allow-all-egress-by-pod", Target: source, Egress: AllowAllByPod},
		}}
}

func DenyAllEgressByPodAllowAllEgressByIP(source *NetpolTarget) *conflictCase {
	return &conflictCase{
		Description: "deny all by pod + allow all by IP from same source",
		Tags:        []string{TagEgress},
		Policies: []*Netpol{
			{Name: "deny-all-egress-by-pod", Target: source, Egress: DenyAllByPod},
			{Name: "allow-all-egress-by-ip", Target: source, Egress: AllowAllByIP},
		}}
}

func DenyAllIngressAllowAllIngressByPod(dest *NetpolTarget) *conflictCase {
	return &conflictCase{
		Description: "deny all + allow all by pod to same source",
		Tags:        []string{TagDenyAll, TagIngress, TagAllPods, TagAllNamespaces},
		Policies: []*Netpol{
			{Name: "deny-all-ingress", Target: dest, Ingress: DenyAll},
			{Name: "allow-all-ingress-by-pod", Target: dest, Ingress: AllowAllByPod},
		}}
}

func DenyAllIngressAllowAllIngressByIP(dest *NetpolTarget) *conflictCase {
	return &conflictCase{
		Description: "deny all + allow all by IP to same source",
		Tags:        []string{TagDenyAll, TagIngress},
		Policies: []*Netpol{
			{Name: "deny-all-ingress", Target: dest, Ingress: DenyAll},
			{Name: "allow-all-ingress-by-ip", Target: dest, Ingress: AllowAllByIP},
		}}
}

func DenyAllIngressByIPAllowAllIngressByPod(dest *NetpolTarget) *conflictCase {
	return &conflictCase{
		Description: "deny all by IP + allow all by pod to same source",
		Tags:        []string{TagIngress, TagAllPods, TagAllNamespaces},
		Policies: []*Netpol{
			{Name: "deny-all-ingress-by-ip", Target: dest, Ingress: DenyAllByIP},
			{Name: "allow-all-ingress-by-pod", Target: dest, Ingress: AllowAllByPod},
		}}
}

func DenyAllIngressByPodAllowAllIngressByIP(dest *NetpolTarget) *conflictCase {
	return &conflictCase{
		Description: "deny all by pod + allow all by IP to same source",
		Tags:        []string{TagIngress},
		Policies: []*Netpol{
			{Name: "deny-all-ingress-by-pod", Target: dest, Ingress: DenyAllByPod},
			{Name: "allow-all-ingress-by-ip", Target: dest, Ingress: AllowAllByIP},
		}}
}

func DenyAllEgressByIP(source *NetpolTarget) *conflictCase {
	return &conflictCase{
		Description: "egress: deny all by IP",
		Tags:        []string{TagEgress},
		Policies: []*Netpol{
			{Name: "deny-all-egress-by-ip", Target: source, Egress: DenyAllByIP},
		}}
}

func DenyAllEgressByPod(source *NetpolTarget) *conflictCase {
	return &conflictCase{
		Description: "egress: deny all by pod",
		Tags:        []string{TagEgress},
		Policies: []*Netpol{
			{Name: "deny-all-egress-by-ip", Target: source, Egress: DenyAllByPod},
		}}
}

func DenyAllIngressByIP(dest *NetpolTarget) *conflictCase {
	return &conflictCase{
		Description: "ingress: deny all by IP",
		Tags:        []string{TagIngress},
		Policies: []*Netpol{
			{Name: "deny-all-ingress-by-ip", Target: dest, Ingress: DenyAllByIP},
		}}
}

func DenyAllIngressByPod(dest *NetpolTarget) *conflictCase {
	return &conflictCase{
		Description: "ingress: deny all by pod",
		Tags:        []string{TagIngress},
		Policies: []*Netpol{
			{Name: "deny-all-ingress-by-ip", Target: dest, Ingress: DenyAllByPod},
		}}
}

func (t *TestCaseGenerator) ConflictTestCases() []*TestCase {
	source := NewNetpolTarget("x", map[string]string{"pod": "b"}, nil)
	destination := NewNetpolTarget("y", map[string]string{"pod": "c"}, nil)
	return t.ConflictNetworkPolicies(source, destination)
}

func (t *TestCaseGenerator) ConflictNetworkPolicies(source *NetpolTarget, dest *NetpolTarget) []*TestCase {
	policySlices := []*conflictCase{
		AllowAllIngressDenyAllEgress(source, dest),
		AllowAllEgressDenyAllIngress(source, dest),

		DenyAllEgressAllowAllEgress(source),
		DenyAllIngressAllowAllIngress(dest),

		DenyAllEgressAllowAllEgressByPod(source),
		DenyAllEgressAllowAllEgressByIP(source),
		DenyAllEgressByIPAllowAllEgressByPod(source),
		DenyAllEgressByPodAllowAllEgressByIP(source),

		DenyAllIngressAllowAllIngressByPod(source),
		DenyAllIngressAllowAllIngressByIP(source),
		DenyAllIngressByIPAllowAllIngressByPod(source),
		DenyAllIngressByPodAllowAllIngressByIP(source),

		DenyAllEgressByIP(source),
		DenyAllEgressByPod(source),

		DenyAllIngressByIP(source),
		DenyAllIngressByPod(source),
	}

	var testCases []*TestCase
	for _, testCase := range policySlices {
		actions := make([]*Action, len(testCase.Policies))
		hasEgress := false
		for i, pol := range testCase.Policies {
			if pol.Egress != nil {
				hasEgress = true
			}
			actions[i] = CreatePolicy(pol.NetworkPolicy())
		}
		if hasEgress && t.AllowDNS {
			actions = append(actions, CreatePolicy(AllowDNSPolicy(source).NetworkPolicy()))
		}
		tags := NewStringSet(testCase.Tags...)
		tags.Add(TagConflict)
		testCases = append(testCases,
			NewSingleStepTestCase(testCase.Description, tags, ProbeAllAvailable, actions...))
	}

	return testCases
}
