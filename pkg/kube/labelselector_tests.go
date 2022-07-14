package kube

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RunLabelSelectorTests() {
	Describe("LabelSelectors", func() {
		It("Should not match empty labels", func() {
			Expect(IsLabelsMatchLabelSelector(map[string]string{}, metav1.LabelSelector{
				MatchLabels: map[string]string{"pod": "b"},
			})).To(BeFalse())
		})
	})
}
