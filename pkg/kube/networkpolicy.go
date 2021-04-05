package kube

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func NetworkPoliciesToTable(policies []*NetworkPolicy) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Policy", "Target", "Direction", "Peer", "Port/Protocol"})

	for _, policy := range policies {
		name := fmt.Sprintf("%s/%s", policy.Namespace, policy.Name)
		//target := SerializeLabelSelector(policy.Spec.PodSelector)
		target := LabelSelectorTableLines(policy.Spec.PodSelector)

		for _, policyType := range policy.Spec.PolicyTypes {
			if policyType == PolicyTypeIngress {
				if len(policy.Spec.Ingress) == 0 {
					table.Append([]string{name, target, "ingress", "none", "none"})
				} else {
					for _, i := range policy.Spec.Ingress {
						table.Append([]string{name, target, "ingress", PrintPeers(i.From), PrintPorts(i.Ports)})
					}
				}
			} else if policyType == PolicyTypeEgress {
				if len(policy.Spec.Egress) == 0 {
					table.Append([]string{name, target, "egress", "none", "none"})
				} else {
					for _, e := range policy.Spec.Egress {
						table.Append([]string{name, target, "egress", PrintPeers(e.To), PrintPorts(e.Ports)})
					}
				}
			}
		}

		table.Append([]string{"", "", "", "", ""})
	}

	table.Render()
	return tableString.String()
}

func PrintPeers(npPeers []NetworkPolicyPeer) string {
	if len(npPeers) == 0 {
		return "all peers"
	}
	var lines []string
	for _, peer := range npPeers {
		if peer.IPBlock != nil {
			lines = append(lines, PrintIPBlock(*peer.IPBlock))
		} else {
			lines = append(lines, PrintNSPodPeer(peer.NamespaceSelector, peer.PodSelector))
		}
	}
	return strings.Join(lines, "\n\n")
}

func PrintIPBlock(i IPBlock) string {
	return fmt.Sprintf("%s except [%s]", i.CIDR, strings.Join(i.Except, ","))
}

func PrintNSPodPeer(nsSelector *metav1.LabelSelector, podSelector *metav1.LabelSelector) string {
	var ns, pod string
	if nsSelector == nil {
		ns = "nil"
	} else {
		ns = SerializeLabelSelector(*nsSelector)
	}
	if podSelector == nil {
		pod = "nil"
	} else {
		pod = SerializeLabelSelector(*podSelector)
	}
	return fmt.Sprintf("ns/pod selector:\n - ns: %s\n - pod: %s", ns, pod)
}

func PrintPorts(npPorts []NetworkPolicyPort) string {
	if len(npPorts) == 0 {
		return "all ports, all protocols"
	}
	var lines []string
	for _, pp := range npPorts {
		var port, protocol string
		if pp.Port == nil {
			port = "all ports"
		} else {
			port = fmt.Sprintf("port %s", pp.Port.String())
		}
		if pp.Protocol == nil {
			protocol = "TCP"
		} else {
			protocol = string(*pp.Protocol)
		}
		if pp.EndPort == nil {
			lines = append(lines, fmt.Sprintf("%s on %s", port, protocol))
		} else {
			lines = append(lines, fmt.Sprintf("[%s, %d] on %s", pp.Port.String(), *pp.EndPort, protocol))
		}
	}
	return strings.Join(lines, "\n")
}
