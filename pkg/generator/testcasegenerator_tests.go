package generator

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func RunTestCaseGeneratorTests() {
	Describe("TestCaseGenerator", func() {
		It("Overall number of test cases", func() {
			cases := NewTestCaseGeneratorReplacement(true, "1.2.3.4", []string{}, []string{}).GenerateTestCases()
			Expect(len(cases)).To(Equal(399))
		})
	})
}
