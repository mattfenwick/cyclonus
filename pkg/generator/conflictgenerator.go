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
		Peers: []networkingv1.NetworkPolicyPeer{
			{
				IPBlock: &networkingv1.IPBlock{
					CIDR: "0.0.0.0/31",
				},
			},
		},
	}

	DenyAllByIP = &NetpolPeers{
		Rules: []*Rule{DenyAllByIPRule},
	}

	DenyAllByPodRule = &Rule{
		Peers: []networkingv1.NetworkPolicyPeer{
			{
				PodSelector: nil,
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"this-will-never-happen": "qrs123"},
				},
			},
		},
	}

	DenyAllByPod = &NetpolPeers{
		Rules: []*Rule{DenyAllByPodRule},
	}
)

func AllowAllIngressDenyAllEgress(source *NetpolTarget, dest *NetpolTarget) []*Netpol {
	return []*Netpol{
		{
			Name:   "deny-all-egress",
			Target: source,
			Egress: DenyAll,
		},
		{
			Name:    "allow-all-ingress",
			Target:  dest,
			Ingress: ExplicitAllowAll,
		},
	}
}

func AllowAllEgressDenyAllIngress(source *NetpolTarget, dest *NetpolTarget) []*Netpol {
	return []*Netpol{
		{
			Name:   "allow-all-egress",
			Target: source,
			Egress: ExplicitAllowAll,
		},
		{
			Name:    "deny-all-ingress",
			Target:  dest,
			Ingress: DenyAll,
		},
	}
}

func DenyAllEgressAllowAllEgress(source *NetpolTarget) []*Netpol {
	return []*Netpol{
		{
			Name:   "deny-all-egress",
			Target: source,
			Egress: DenyAll,
		},
		{
			Name:   "allow-all-egress",
			Target: source,
			Egress: ExplicitAllowAll,
		},
	}
}

func DenyAllIngressAllowAllIngress(dest *NetpolTarget) []*Netpol {
	return []*Netpol{
		{Name: "deny-all-ingress", Target: dest, Ingress: DenyAll},
		{Name: "allow-all-ingress", Target: dest, Ingress: ExplicitAllowAll},
	}
}

func DenyAllEgressAllowAllEgressByPod(source *NetpolTarget) []*Netpol {
	return []*Netpol{
		{Name: "deny-all-egress", Target: source, Egress: DenyAll},
		{Name: "allow-all-egress-by-pod", Target: source, Egress: AllowAllByPod},
	}
}

func DenyAllEgressAllowAllEgressByIP(source *NetpolTarget) []*Netpol {
	return []*Netpol{
		{Name: "deny-all-egress", Target: source, Egress: DenyAll},
		{Name: "allow-all-egress-by-ip", Target: source, Egress: AllowAllByIP},
	}
}

func DenyAllEgressByIPAllowAllEgressByPod(source *NetpolTarget) []*Netpol {
	return []*Netpol{
		{Name: "deny-all-egress-by-ip", Target: source, Egress: DenyAllByIP},
		{Name: "allow-all-egress-by-pod", Target: source, Egress: AllowAllByPod},
	}
}

func DenyAllEgressByPodAllowAllEgressByIP(source *NetpolTarget) []*Netpol {
	return []*Netpol{
		{Name: "deny-all-egress-by-pod", Target: source, Egress: DenyAllByPod},
		{Name: "allow-all-egress-by-ip", Target: source, Egress: AllowAllByIP},
	}
}

func DenyAllIngressAllowAllIngressByPod(source *NetpolTarget) []*Netpol {
	return []*Netpol{
		{Name: "deny-all-ingress", Target: source, Ingress: DenyAll},
		{Name: "allow-all-ingress-by-pod", Target: source, Ingress: AllowAllByPod},
	}
}

func DenyAllIngressAllowAllIngressByIP(source *NetpolTarget) []*Netpol {
	return []*Netpol{
		{Name: "deny-all-ingress", Target: source, Ingress: DenyAll},
		{Name: "allow-all-ingress-by-ip", Target: source, Ingress: AllowAllByIP},
	}
}

func DenyAllIngressByIPAllowAllIngressByPod(source *NetpolTarget) []*Netpol {
	return []*Netpol{
		{Name: "deny-all-ingress-by-ip", Target: source, Ingress: DenyAllByIP},
		{Name: "allow-all-ingress-by-pod", Target: source, Ingress: AllowAllByPod},
	}
}

func DenyAllIngressByPodAllowAllIngressByIP(source *NetpolTarget) []*Netpol {
	return []*Netpol{
		{Name: "deny-all-ingress-by-pod", Target: source, Ingress: DenyAllByPod},
		{Name: "allow-all-ingress-by-ip", Target: source, Ingress: AllowAllByIP},
	}
}

func DenyAllEgressByIP(source *NetpolTarget) []*Netpol {
	return []*Netpol{
		{Name: "deny-all-egress-by-ip", Target: source, Egress: DenyAllByIP},
	}
}

func DenyAllEgressByPod(source *NetpolTarget) []*Netpol {
	return []*Netpol{
		{Name: "deny-all-egress-by-ip", Target: source, Egress: DenyAllByPod},
	}
}

func DenyAllIngressByIP(source *NetpolTarget) []*Netpol {
	return []*Netpol{
		{Name: "deny-all-ingress-by-ip", Target: source, Ingress: DenyAllByIP},
	}
}

func DenyAllIngressByPod(source *NetpolTarget) []*Netpol {
	return []*Netpol{
		{Name: "deny-all-ingress-by-ip", Target: source, Ingress: DenyAllByPod},
	}
}

type ConflictGenerator struct {
	AllowDNS    bool
	Source      *NetpolTarget
	Destination *NetpolTarget
}

func (c *ConflictGenerator) GenerateTestCases() []*TestCase {
	return c.NetworkPolicies(c.Source, c.Destination)
}

func (c *ConflictGenerator) NetworkPolicies(source *NetpolTarget, dest *NetpolTarget) []*TestCase {
	policySlices := [][]*Netpol{
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
	for _, slice := range policySlices {
		actions := make([]*Action, len(slice))
		hasEgress := false
		for i, pol := range slice {
			if pol.Egress != nil {
				hasEgress = true
			}
			actions[i] = CreatePolicy(pol.NetworkPolicy())
		}
		if hasEgress && c.AllowDNS {
			actions = append(actions, CreatePolicy(AllowDNSPolicy(source).NetworkPolicy()))
		}
		testCases = append(testCases, NewTestCase(actions))
	}

	return testCases
}
