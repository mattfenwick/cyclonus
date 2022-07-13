package linter

import (
	"fmt"
	collections "github.com/mattfenwick/collections/pkg"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/olekukonko/tablewriter"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"strings"
)

/*
warnings:
 - dns not allowed
 - target isolated, nothing allowed
 - pod not hit by any target (i.e. implicit allow-all)
 - explicit allow-all from target
*/

type Check string

const (
	// omitting the namespace will create the policy in the default namespace
	CheckSourceMissingNamespace Check = "CheckSourceMissingNamespace"
	// omitting the protocol from a NetworkPolicyPort will default to TCP
	CheckSourcePortMissingProtocol Check = "CheckSourcePortMissingProtocol"
	// omitting the types can sometimes be automatically handled; but it's better to explicitly list them
	CheckSourceMissingPolicyTypes Check = "CheckSourceMissingPolicyTypes"
	// if the policy has ingress/egress rules, then the corresponding type should be present
	CheckSourceMissingPolicyTypeIngress Check = "CheckSourceMissingPolicyTypeIngress"
	CheckSourceMissingPolicyTypeEgress  Check = "CheckSourceMissingPolicyTypeEgress"
	// duplicate names
	CheckSourceDuplicatePolicyName Check = "CheckSourceDuplicatePolicyName"

	CheckDNSBlockedOnTCP         Check = "CheckDNSBlockedOnTCP"
	CheckDNSBlockedOnUDP         Check = "CheckDNSBlockedOnUDP"
	CheckTargetAllIngressBlocked Check = "CheckTargetAllIngressBlocked"
	CheckTargetAllEgressBlocked  Check = "CheckTargetAllEgressBlocked"
	CheckTargetAllIngressAllowed Check = "CheckTargetAllIngressAllowed"
	CheckTargetAllEgressAllowed  Check = "CheckTargetAllEgressAllowed"

	// TODO add check that rule is unnecessary b/c another rule exactly supercedes it
)

func (a Check) Equal(b Check) bool {
	// TODO why is this necessary?  why can't we use existing String implementation?
	return a == b
}

type Warning struct {
	Check        Check
	Target       *matcher.Target
	SourcePolicy *networkingv1.NetworkPolicy
}

func WarningsTable(warnings []*Warning) string {
	str := &strings.Builder{}
	table := tablewriter.NewWriter(str)
	table.SetHeader([]string{"Source/Resolved", "Type", "Target", "Source Policies"})
	table.SetRowLine(true)
	table.SetReflowDuringAutoWrap(false)
	table.SetAutoWrapText(false)

	for _, warning := range warnings {
		if warning.SourcePolicy != nil {
			p := warning.SourcePolicy
			table.Append([]string{"Source", string(warning.Check), "", p.Namespace + "/" + p.Name})
		} else {
			t := warning.Target
			var source []string
			for _, policy := range t.SourceRules {
				source = append(source, policy.Namespace+"/"+policy.Name)
			}
			target := fmt.Sprintf("namespace: %s\n\npod selector:\n%s", t.Namespace, utils.YamlString(t.PodSelector))
			table.Append([]string{"Resolved", string(warning.Check), target, strings.Join(source, "\n")})
		}
	}

	table.Render()
	return str.String()
}

func Lint(kubePolicies []*networkingv1.NetworkPolicy, skip *collections.Set[Check]) []*Warning {
	policies := matcher.BuildNetworkPolicies(false, kubePolicies)
	warnings := append(LintSourcePolicies(kubePolicies), LintResolvedPolicies(policies)...)

	// TODO do some stuff with comparing simplified to non-simplified policies

	var filtered []*Warning
	for _, warning := range warnings {
		if !skip.Contains(warning.Check) {
			filtered = append(filtered, warning)
		}
	}
	return filtered
}

func LintSourcePolicies(kubePolicies []*networkingv1.NetworkPolicy) []*Warning {
	var ws []*Warning
	names := map[string]map[string]bool{}
	for _, policy := range kubePolicies {
		ns, name := policy.Namespace, policy.Name
		if _, ok := names[ns]; !ok {
			names[ns] = map[string]bool{}
		}
		if names[ns][name] {
			ws = append(ws, &Warning{Check: CheckSourceDuplicatePolicyName, SourcePolicy: policy})
		}
		names[ns][name] = true

		if ns == "" {
			ws = append(ws, &Warning{Check: CheckSourceMissingNamespace, SourcePolicy: policy})
		}

		if len(policy.Spec.PolicyTypes) == 0 {
			ws = append(ws, &Warning{Check: CheckSourceMissingPolicyTypes, SourcePolicy: policy})
		}

		ingress, egress := false, false
		for _, policyType := range policy.Spec.PolicyTypes {
			switch policyType {
			case networkingv1.PolicyTypeEgress:
				egress = true
			case networkingv1.PolicyTypeIngress:
				ingress = true
			}
		}
		if len(policy.Spec.Ingress) > 0 && !ingress {
			ws = append(ws, &Warning{Check: CheckSourceMissingPolicyTypeIngress, SourcePolicy: policy})
		}
		if len(policy.Spec.Egress) > 0 && !egress {
			ws = append(ws, &Warning{Check: CheckSourceMissingPolicyTypeEgress, SourcePolicy: policy})
		}

		for _, ingressRule := range policy.Spec.Ingress {
			ws = append(ws, LintNetworkPolicyPorts(policy, ingressRule.Ports)...)
		}
		for _, egressRule := range policy.Spec.Egress {
			ws = append(ws, LintNetworkPolicyPorts(policy, egressRule.Ports)...)
		}
	}
	return ws
}

func LintNetworkPolicyPorts(policy *networkingv1.NetworkPolicy, ports []networkingv1.NetworkPolicyPort) []*Warning {
	var ws []*Warning
	for _, port := range ports {
		if port.Protocol == nil {
			ws = append(ws, &Warning{Check: CheckSourcePortMissingProtocol, SourcePolicy: policy})
		}
	}
	return ws
}

func LintResolvedPolicies(policies *matcher.Policy) []*Warning {
	var ws []*Warning
	for _, egress := range policies.Egress {
		if !egress.Allows(&matcher.TrafficPeer{Internal: nil, IP: "8.8.8.8"}, 53, "", v1.ProtocolTCP) {
			ws = append(ws, &Warning{Check: CheckDNSBlockedOnTCP, Target: egress})
		}
		if !egress.Allows(&matcher.TrafficPeer{Internal: nil, IP: "8.8.8.8"}, 53, "", v1.ProtocolUDP) {
			ws = append(ws, &Warning{Check: CheckDNSBlockedOnUDP, Target: egress})
		}

		if len(egress.Peers) == 0 {
			ws = append(ws, &Warning{Check: CheckTargetAllEgressBlocked, Target: egress})
		}
		for _, peer := range egress.Peers {
			if _, ok := peer.(*matcher.PortsForAllPeersMatcher); ok {
				ws = append(ws, &Warning{Check: CheckTargetAllEgressAllowed, Target: egress})
			}
		}
	}

	for _, ingress := range policies.Ingress {
		if len(ingress.Peers) == 0 {
			ws = append(ws, &Warning{Check: CheckTargetAllIngressBlocked, Target: ingress})
		}
		for _, peer := range ingress.Peers {
			if _, ok := peer.(*matcher.PortsForAllPeersMatcher); ok {
				ws = append(ws, &Warning{Check: CheckTargetAllIngressAllowed, Target: ingress})
			}
		}

	}

	return ws
}
