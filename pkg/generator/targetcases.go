package generator

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (t *TestCaseGenerator) TargetTestCases() []*TestCase {
	var cases []*TestCase

	// TODO want to test the empty-string-to-default-namespace behavior, but the kube client doesn't allow an empty
	//   namespace like kubectl does
	//cases = append(cases, NewSingleStepTestCase("set namespace to empty string", NewStringSet(TagTargetNamespace), ProbeAllAvailable,
	//	CreatePolicy(BuildPolicy(SetNamespace("")).NetworkPolicy())))

	for _, ns := range t.Namespaces {
		tags := NewStringSet(TagTargetNamespace)
		cases = append(cases, NewSingleStepTestCase(fmt.Sprintf("set namespace to %s", ns), tags, ProbeAllAvailable,
			CreatePolicy(BuildPolicy(SetNamespace(ns)).NetworkPolicy())))
	}

	for _, selector := range []metav1.LabelSelector{*emptySelector, *podAMatchLabelsSelector, *podABMatchExpressionsSelector} {
		cases = append(cases, NewSingleStepTestCase(
			fmt.Sprintf("set pod selector to %s", kube.SerializeLabelSelector(selector)),
			NewStringSet(TagTargetPodSelector),
			ProbeAllAvailable,
			CreatePolicy(BuildPolicy(SetPodSelector(selector)).NetworkPolicy())))
	}
	return cases
}
