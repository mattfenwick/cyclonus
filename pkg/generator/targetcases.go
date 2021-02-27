package generator

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (t *TestCaseGeneratorReplacement) TargetTestCases() []*TestCase {
	var cases []*TestCase
	for _, ns := range t.Namespaces {
		tags := NewStringSet(TagTargetNamespace)
		cases = append(cases, NewSingleStepTestCase(fmt.Sprintf("set namespace to %s", ns), tags, ProbeAllAvailable,
			CreatePolicy(BuildPolicy(SetNamespace(ns)).NetworkPolicy())))
	}

	for _, selector := range []metav1.LabelSelector{*emptySelector, *podAMatchLabelsSelector, *podABMatchExpressionsSelector} {
		tags := NewStringSet(TagTargetPodSelector)
		cases = append(cases, NewSingleStepTestCase(
			fmt.Sprintf("set pod selector to %s", kube.SerializeLabelSelector(selector)),
			tags,
			ProbeAllAvailable,
			CreatePolicy(BuildPolicy(SetPodSelector(selector)).NetworkPolicy())))
	}
	return cases
}
