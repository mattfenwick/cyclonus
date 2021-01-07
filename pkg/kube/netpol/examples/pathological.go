package examples

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	LabelsAB = map[string]string{"a": "b"}
	LabelsCD = map[string]string{"b": "d"}
	LabelsEF = map[string]string{"e": "f"}
	LabelsGH = map[string]string{"g": "g"}

	SelectorAB = &metav1.LabelSelector{
		MatchLabels: LabelsAB,
	}
	SelectorCD = &metav1.LabelSelector{
		MatchLabels: LabelsCD,
	}
	SelectorEF = &metav1.LabelSelector{
		MatchLabels: LabelsEF,
	}
	SelectorGH = &metav1.LabelSelector{
		MatchLabels: LabelsGH,
	}

	Namespace = "pathological-namespace"
)

// allow nothing (i.e. deny all)

var AllowNoIngress = &networkingv1.NetworkPolicy{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "allow-no-ingress",
		Namespace: Namespace,
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
	},
}

var AllowNoIngress_EmptyIngress = &networkingv1.NetworkPolicy{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "allow-no-ingress-empty-ingress",
		Namespace: Namespace,
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{},
		Ingress:     []networkingv1.NetworkPolicyIngressRule{},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
	},
}

var AllowNoEgress = &networkingv1.NetworkPolicy{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "allow-no-egress",
		Namespace: Namespace,
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
	},
}

var AllowNoEgress_EmptyEgress = &networkingv1.NetworkPolicy{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "allow-no-egress-empty-egress",
		Namespace: Namespace,
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{},
		Egress:      []networkingv1.NetworkPolicyEgressRule{},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
	},
}

var AllowNoIngressAllowNoEgress = &networkingv1.NetworkPolicy{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "allow-no-ingress-allow-no-egress",
		Namespace: Namespace,
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress, networkingv1.PolicyTypeIngress},
	},
}

var AllowNoIngressAllowNoEgress_EmptyEgressEmptyIngress = &networkingv1.NetworkPolicy{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "allow-no-ingress-allow-no-egress-empty-egress-empty-ingress",
		Namespace: Namespace,
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{},
		Ingress:     []networkingv1.NetworkPolicyIngressRule{},
		Egress:      []networkingv1.NetworkPolicyEgressRule{},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress, networkingv1.PolicyTypeIngress},
	},
}

// allow all

var AllowAllIngress = &networkingv1.NetworkPolicy{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "allow-all-ingress",
		Namespace: Namespace,
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		Ingress: []networkingv1.NetworkPolicyIngressRule{
			{},
		},
	},
}

var AllowAllEgress = &networkingv1.NetworkPolicy{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "allow-all-egress",
		Namespace: Namespace,
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
		Egress: []networkingv1.NetworkPolicyEgressRule{
			{},
		},
	},
}

var AllowAllIngressAllowAllEgress = &networkingv1.NetworkPolicy{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "allow-all-ingress-allow-all-egress",
		Namespace: Namespace,
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress, networkingv1.PolicyTypeIngress},
		Egress: []networkingv1.NetworkPolicyEgressRule{
			{},
		},
		Ingress: []networkingv1.NetworkPolicyIngressRule{
			{},
		},
	},
}

// allow based on matching pod selector and namespace selector

var AllowAllPodsInPolicyNamespacePeer = networkingv1.NetworkPolicyPeer{
	PodSelector:       nil,
	NamespaceSelector: nil,
}

var AllowAllPodsInAllNamespacesPeer = networkingv1.NetworkPolicyPeer{
	PodSelector:       nil,
	NamespaceSelector: &metav1.LabelSelector{},
}

var AllowAllPodsInMatchingNamespacesPeer = networkingv1.NetworkPolicyPeer{
	PodSelector:       nil,
	NamespaceSelector: SelectorAB,
}

var AllowAllPodsInPolicyNamespacePeer_EmptyPodSelector = networkingv1.NetworkPolicyPeer{
	PodSelector:       &metav1.LabelSelector{},
	NamespaceSelector: nil,
}

var AllowAllPodsInAllNamespacesPeer_EmptyPodSelector = networkingv1.NetworkPolicyPeer{
	PodSelector:       &metav1.LabelSelector{},
	NamespaceSelector: &metav1.LabelSelector{},
}

var AllowAllPodsInMatchingNamespacesPeer_EmptyPodSelector = networkingv1.NetworkPolicyPeer{
	PodSelector:       &metav1.LabelSelector{},
	NamespaceSelector: SelectorAB,
}

var AllowMatchingPodsInPolicyNamespacePeer = networkingv1.NetworkPolicyPeer{
	PodSelector:       SelectorCD,
	NamespaceSelector: nil,
}

var AllowMatchingPodsInAllNamespacesPeer = networkingv1.NetworkPolicyPeer{
	PodSelector:       SelectorEF,
	NamespaceSelector: &metav1.LabelSelector{},
}

var AllowMatchingPodsInMatchingNamespacesPeer = networkingv1.NetworkPolicyPeer{
	PodSelector:       SelectorGH,
	NamespaceSelector: SelectorAB,
}

var AllowIPBlockPeer = networkingv1.NetworkPolicyPeer{
	IPBlock: &networkingv1.IPBlock{
		CIDR: "10.0.0.1/24",
		Except: []string{
			"10.0.0.2",
		},
	},
}

// allow based on matching port and protocol

var AllowAllPortsOnProtocol = networkingv1.NetworkPolicyPort{
	Protocol: &SCTP,
	Port:     nil,
}

var AllowNumberedPortOnProtocol = networkingv1.NetworkPolicyPort{
	Protocol: &TCP,
	Port:     &Port9001Ref,
}

var AllowNamedPortOnProtocol = networkingv1.NetworkPolicyPort{
	Protocol: &UDP,
	Port:     &PortHello,
}
