package netpol

import (
	"k8s.io/apimachinery/pkg/util/intstr"
)

type NetworkPolicySpec struct {
	PodSelector LabelSelector              // metav1.LabelSelector
	Ingress     []NetworkPolicyIngressRule // networkingv1.NetworkPolicyIngressRule
	Egress      []NetworkPolicyEgressRule  // networkingv1.NetworkPolicyEgressRule
	PolicyTypes []PolicyType               // networkingv1.PolicyType
}

type NetworkPolicyIngressRule struct {
	Ports []NetworkPolicyPort // networkingv1.NetworkPolicyPort
	From  []NetworkPolicyPeer // networkingv1.NetworkPolicyPeer
}

type NetworkPolicyPort struct {
	Protocol *Protocol // v1.Protocol
	Port     *intstr.IntOrString
}

type NetworkPolicyPeer struct {
	PodSelector       *LabelSelector // metav1.LabelSelector
	NamespaceSelector *LabelSelector // metav1.LabelSelector
	IPBlock           *IPBlock       // networkingv1.IPBlock
}

type NetworkPolicyEgressRule struct {
	Ports []NetworkPolicyPort
	To    []NetworkPolicyPeer
}

type IPBlock struct {
	CIDR   string
	Except []string
}

type Protocol string

const (
	ProtocolTCP  Protocol = "TCP"
	ProtocolUDP  Protocol = "UDP"
	ProtocolSCTP Protocol = "SCTP"
)

type PolicyType string

const (
	PolicyTypeIngress PolicyType = "Ingress"
	PolicyTypeEgress  PolicyType = "Egress"
)

type LabelSelector struct {
	MatchLabels      map[string]string
	MatchExpressions []LabelSelectorRequirement // metav1.LabelSelectorRequirement
}

type LabelSelectorRequirement struct {
	Key      string
	Operator LabelSelectorOperator // metav1.LabelSelectorOperator
	Values   []string
}

type LabelSelectorOperator string

const (
	LabelSelectorOpIn           LabelSelectorOperator = "In"
	LabelSelectorOpNotIn        LabelSelectorOperator = "NotIn"
	LabelSelectorOpExists       LabelSelectorOperator = "Exists"
	LabelSelectorOpDoesNotExist LabelSelectorOperator = "DoesNotExist"
)
