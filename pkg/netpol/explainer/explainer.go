package explainer

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube/netpol/examples"
	log "github.com/sirupsen/logrus"
	networkingv1 "k8s.io/api/networking/v1"
	"strings"
)

type ExplanationTarget struct {
	Namespace string
	Pods      string
}

type ExplanationRule struct {
	Pods       string
	Namespaces string
}

type ExplanationEgress struct {
	Ports []string
	Rules []*ExplanationRule
}

func (ee *ExplanationEgress) PrettyPrintPorts() string {
	var ports []string
	if len(ee.Ports) == 0 {
		ports = []string{"all allowed"}
	} else {
		for _, p := range ee.Ports {
			ports = append(ports, p)
		}
	}
	return "ports: " + strings.Join(ports, " OR ")
}

func (ee *ExplanationEgress) PrettyPrintRules() string {
	if len(ee.Rules) == 0 {
		return "all pods, namespaces, and ips are allowed"
	}
	var allowed []string
	for _, rule := range ee.Rules {
		allowed = append(allowed, fmt.Sprintf("(%s AND %s)", rule.Namespaces, rule.Pods))
	}
	return strings.Join(allowed, " OR ")
}

func (ee *ExplanationEgress) PrettyPrint() string {
	return fmt.Sprintf("(%s) AND (%s)", ee.PrettyPrintPorts(), ee.PrettyPrintRules())
}

type ExplanationIngress struct {
	Ports []string
	Rules []*ExplanationRule
}

func (ei *ExplanationIngress) PrettyPrintPorts() string {
	var ports []string
	if len(ei.Ports) == 0 {
		ports = []string{"all allowed"}
	} else {
		for _, p := range ei.Ports {
			ports = append(ports, p)
		}
	}
	return "ports: " + strings.Join(ports, " OR ")
}

func (ei *ExplanationIngress) PrettyPrintRules() string {
	if len(ei.Rules) == 0 {
		return "all pods, namespaces, and ips are allowed"
	}
	var allowed []string
	for _, rule := range ei.Rules {
		allowed = append(allowed, fmt.Sprintf("(%s AND %s)", rule.Namespaces, rule.Pods))
	}
	return strings.Join(allowed, " OR ")
}

func (ei *ExplanationIngress) PrettyPrint() string {
	return fmt.Sprintf("(%s) AND (%s)", ei.PrettyPrintPorts(), ei.PrettyPrintRules())
}

type Explanation struct {
	Target    *ExplanationTarget
	Ingress   []*ExplanationIngress
	Egress    []*ExplanationEgress
	IsIngress bool
	IsEgress  bool
}

func (e *Explanation) PrettyPrint() string {
	out := []string{
		"target namespace: " + e.Target.Namespace,
		"target pods: " + e.Target.Pods,
	}
	if e.IsIngress {
		out = append(out, "ingress:")
		if len(e.Ingress) == 0 {
			out = append(out, " - no ingress is allowed")
		} else {
			var ingresses []string
			for _, ingress := range e.Ingress {
				ingresses = append(ingresses, " - "+ingress.PrettyPrint())
			}
			out = append(out, strings.Join(ingresses, " OR "))
		}
	}

	if e.IsEgress {
		out = append(out, "egress:")
		if len(e.Egress) == 0 {
			out = append(out, " - no egress is allowed")
		} else {
			var egresses []string
			for _, egress := range e.Egress {
				egresses = append(egresses, " - "+egress.PrettyPrint())
			}
			out = append(out, strings.Join(egresses, " OR "))
		}
	}
	return strings.Join(out, "\n")
}

func ExplainPolicy(policy *networkingv1.NetworkPolicy) *Explanation {
	log.Warnf("TODO policy.spec.podselector.MatchExpressions is currently ignored")

	targetLabels := policy.Spec.PodSelector.MatchLabels
	var targetPods string
	if len(targetLabels) == 0 {
		targetPods = "all pods"
	} else {
		targetPods = fmt.Sprintf("pods with labels %s", examples.LabelString(targetLabels))
	}

	isIngress, isEgress := false, false
	for _, pType := range policy.Spec.PolicyTypes {
		switch pType {
		case networkingv1.PolicyTypeIngress:
			isIngress = true
		case networkingv1.PolicyTypeEgress:
			isEgress = true
		}
	}

	return &Explanation{
		Target: &ExplanationTarget{
			Namespace: policy.Namespace,
			Pods:      targetPods,
		},
		Ingress:   ExplainIngress(policy),
		Egress:    ExplainEgress(policy),
		IsIngress: isIngress,
		IsEgress:  isEgress,
	}
}

func ExplainEgress(policy *networkingv1.NetworkPolicy) []*ExplanationEgress {
	var egresses []*ExplanationEgress
	for _, rule := range policy.Spec.Egress {
		var ports []string
		for _, p := range rule.Ports {
			protocol := "TCP"
			if p.Protocol != nil {
				protocol = string(*p.Protocol)
			}
			ports = append(ports, fmt.Sprintf("%s - %+v", protocol, p.Port))
		}

		var rules []*ExplanationRule
		for _, to := range rule.To {
			log.Warnf("TODO ingress/to/ipblock is currently ignored")
			var pods, ns string

			if to.PodSelector == nil {
				log.Warnf("TODO this is probably worth a test -- when pod selector isn't specified, does that mean 'all pods'?")
				pods = "all pods"
			} else if len(to.PodSelector.MatchLabels) == 0 {
				pods = "all pods"
			} else {
				pods = fmt.Sprintf("pods with labels %s", examples.LabelString(to.PodSelector.MatchLabels))
			}

			if to.NamespaceSelector == nil {
				ns = fmt.Sprintf("ns %s", policy.Namespace)
			} else if len(to.NamespaceSelector.MatchLabels) == 0 {
				ns = "all namespaces"
			} else {
				ns = fmt.Sprintf("namespaces with labels %s", examples.LabelString(to.NamespaceSelector.MatchLabels))
			}

			rules = append(rules, &ExplanationRule{
				Pods:       pods,
				Namespaces: ns,
			})
		}
		egresses = append(egresses, &ExplanationEgress{Ports: ports, Rules: rules})
	}
	return egresses
}

func ExplainIngress(policy *networkingv1.NetworkPolicy) []*ExplanationIngress {
	var ingresses []*ExplanationIngress
	for _, rule := range policy.Spec.Ingress {
		var ports []string
		for _, p := range rule.Ports {
			protocol := "TCP"
			if p.Protocol != nil {
				protocol = string(*p.Protocol)
			}
			ports = append(ports, fmt.Sprintf("%s - %+v", protocol, p.Port))
		}

		var rules []*ExplanationRule
		for _, from := range rule.From {
			log.Warnf("TODO ingress/from/ipblock is currently ignored")
			var pods, ns string

			if from.PodSelector == nil {
				log.Warnf("TODO this is probably worth a test -- when pod selector isn't specified, does that mean 'all pods'?")
				pods = "all pods"
			} else if len(from.PodSelector.MatchLabels) == 0 {
				pods = "all pods"
			} else {
				pods = fmt.Sprintf("only pods with labels %s", examples.LabelString(from.PodSelector.MatchLabels))
			}

			if from.NamespaceSelector == nil {
				ns = fmt.Sprintf("only namespace %s", policy.Namespace)
			} else if len(from.NamespaceSelector.MatchLabels) == 0 {
				ns = "all namespaces"
			} else {
				ns = fmt.Sprintf("only namespaces with labels %s", examples.LabelString(from.NamespaceSelector.MatchLabels))
			}

			rules = append(rules, &ExplanationRule{
				Pods:       pods,
				Namespaces: ns,
			})
		}
		ingresses = append(ingresses, &ExplanationIngress{Ports: ports, Rules: rules})
	}
	return ingresses
}
