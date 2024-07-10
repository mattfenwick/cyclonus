package linter

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/set"
	"github.com/mattfenwick/collections/pkg/slice"
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
	// CheckSourceMissingNamespace omitting the namespace will create the policy in the default namespace
	CheckSourceMissingNamespace Check = "CheckSourceMissingNamespace"
	// CheckSourcePortMissingProtocol omitting the protocol from a NetworkPolicyPort will default to TCP
	CheckSourcePortMissingProtocol Check = "CheckSourcePortMissingProtocol"
	// CheckSourceMissingPolicyTypes omitting the types can sometimes be automatically handled; but it's better to explicitly list them
	CheckSourceMissingPolicyTypes Check = "CheckSourceMissingPolicyTypes"
	// CheckSourceMissingPolicyTypeIngress if the policy has ingress rules, then that type should be present
	CheckSourceMissingPolicyTypeIngress Check = "CheckSourceMissingPolicyTypeIngress"
	// CheckSourceMissingPolicyTypeEgress if the policy has egress rules, then that type should be present
	CheckSourceMissingPolicyTypeEgress Check = "CheckSourceMissingPolicyTypeEgress"
	// CheckSourceDuplicatePolicyName duplicate names of source network policies
	CheckSourceDuplicatePolicyName Check = "CheckSourceDuplicatePolicyName"

	CheckDNSBlockedOnTCP         Check = "CheckDNSBlockedOnTCP"
	CheckDNSBlockedOnUDP         Check = "CheckDNSBlockedOnUDP"
	CheckTargetAllIngressBlocked Check = "CheckTargetAllIngressBlocked"
	CheckTargetAllEgressBlocked  Check = "CheckTargetAllEgressBlocked"
	CheckTargetAllIngressAllowed Check = "CheckTargetAllIngressAllowed"
	CheckTargetAllEgressAllowed  Check = "CheckTargetAllEgressAllowed"

	// TODO add check that rule is unnecessary b/c another rule exactly supersedes it
)

type Warning interface {
	OriginIsSource() bool
	GetCheck() Check
	GetTarget() string
	GetSourcePolicies() string
}

type sourceWarning struct {
	Check        Check
	SourcePolicy *networkingv1.NetworkPolicy
}

func (s *sourceWarning) OriginIsSource() bool {
	return true
}

func (s *sourceWarning) GetCheck() Check {
	return s.Check
}

func (s *sourceWarning) GetTarget() string {
	return ""
}

func (s *sourceWarning) GetSourcePolicies() string {
	return NetpolKey(s.SourcePolicy)
}

func NetpolKey(netpol *networkingv1.NetworkPolicy) string {
	return fmt.Sprintf("%s/%s", netpol.Namespace, netpol.Name)
}

func sortKey(w Warning) []string {
	origin := "1"
	if w.OriginIsSource() {
		origin = "0"
	}
	return []string{origin, string(w.GetCheck()), w.GetTarget(), w.GetSourcePolicies()}
}

type resolvedWarning struct {
	Check          Check
	Target         *matcher.Target
	originPolicies string
}

func (r *resolvedWarning) OriginIsSource() bool {
	return false
}

func (r *resolvedWarning) GetCheck() Check {
	return r.Check
}

func (r *resolvedWarning) GetTarget() string {
	return fmt.Sprintf("namespace: %s\n\npod selector:\n%s", r.Target.Namespace, utils.YamlString(r.Target.PodSelector))
}

func (r *resolvedWarning) GetSourcePolicies() string {
	target := slice.Sort(slice.Map(NetpolKey, r.Target.SourceRules))
	return strings.Join(target, "\n")
}

func WarningsTable(warnings []Warning) string {
	str := &strings.Builder{}
	table := tablewriter.NewWriter(str)
	table.SetHeader([]string{"Source/Resolved", "Type", "Target", "Source Policies"})
	table.SetRowLine(true)
	table.SetReflowDuringAutoWrap(false)
	table.SetAutoWrapText(false)

	sortedWarnings := slice.SortOnBy(sortKey, slice.ComparePairwise[string](), warnings)
	for _, w := range sortedWarnings {
		origin := "Source"
		if !w.OriginIsSource() {
			origin = "Resolved"
		}
		table.Append([]string{origin, string(w.GetCheck()), w.GetTarget(), w.GetSourcePolicies()})
	}

	table.Render()
	return str.String()
}

func Lint(kubePolicies []*networkingv1.NetworkPolicy, skip *set.Set[Check]) []Warning {
	policies := matcher.BuildNetworkPolicies(false, kubePolicies)
	warnings := append(LintSourcePolicies(kubePolicies), LintResolvedPolicies(policies)...)

	// TODO do some stuff with comparing simplified to non-simplified policies

	var filtered []Warning
	for _, warning := range warnings {
		if !skip.Contains(warning.GetCheck()) {
			filtered = append(filtered, warning)
		}
	}
	return filtered
}

func LintSourcePolicies(kubePolicies []*networkingv1.NetworkPolicy) []Warning {
	var ws []Warning
	names := map[string]map[string]bool{}
	for _, policy := range kubePolicies {
		ns, name := policy.Namespace, policy.Name
		if _, ok := names[ns]; !ok {
			names[ns] = map[string]bool{}
		}
		if names[ns][name] {
			ws = append(ws, &sourceWarning{Check: CheckSourceDuplicatePolicyName, SourcePolicy: policy})
		}
		names[ns][name] = true

		if ns == "" {
			ws = append(ws, &sourceWarning{Check: CheckSourceMissingNamespace, SourcePolicy: policy})
		}

		if len(policy.Spec.PolicyTypes) == 0 {
			ws = append(ws, &sourceWarning{Check: CheckSourceMissingPolicyTypes, SourcePolicy: policy})
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
			ws = append(ws, &sourceWarning{Check: CheckSourceMissingPolicyTypeIngress, SourcePolicy: policy})
		}
		if len(policy.Spec.Egress) > 0 && !egress {
			ws = append(ws, &sourceWarning{Check: CheckSourceMissingPolicyTypeEgress, SourcePolicy: policy})
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

func LintNetworkPolicyPorts(policy *networkingv1.NetworkPolicy, ports []networkingv1.NetworkPolicyPort) []Warning {
	var ws []Warning
	for _, port := range ports {
		if port.Protocol == nil {
			ws = append(ws, &sourceWarning{Check: CheckSourcePortMissingProtocol, SourcePolicy: policy})
		}
	}
	return ws
}

func LintResolvedPolicies(policies *matcher.Policy) []Warning {
	var ws []Warning
	for _, egress := range policies.Egress {
		if !egress.Allows(&matcher.TrafficPeer{Internal: nil, IP: "8.8.8.8"}, 53, "", v1.ProtocolTCP) {
			ws = append(ws, &resolvedWarning{Check: CheckDNSBlockedOnTCP, Target: egress})
		}
		if !egress.Allows(&matcher.TrafficPeer{Internal: nil, IP: "8.8.8.8"}, 53, "", v1.ProtocolUDP) {
			ws = append(ws, &resolvedWarning{Check: CheckDNSBlockedOnUDP, Target: egress})
		}

		if len(egress.Peers) == 0 {
			ws = append(ws, &resolvedWarning{Check: CheckTargetAllEgressBlocked, Target: egress})
		}
		for _, peer := range egress.Peers {
			if _, ok := peer.(*matcher.PortsForAllPeersMatcher); ok {
				ws = append(ws, &resolvedWarning{Check: CheckTargetAllEgressAllowed, Target: egress})
			}
		}
	}

	for _, ingress := range policies.Ingress {
		if len(ingress.Peers) == 0 {
			ws = append(ws, &resolvedWarning{Check: CheckTargetAllIngressBlocked, Target: ingress})
		}
		for _, peer := range ingress.Peers {
			if _, ok := peer.(*matcher.PortsForAllPeersMatcher); ok {
				ws = append(ws, &resolvedWarning{Check: CheckTargetAllIngressAllowed, Target: ingress})
			}
		}

	}

	return ws
}
