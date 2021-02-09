package generator

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func RunDiscreteGeneratorTests() {
	Describe("DiscreteGenerator", func() {
		It("Complicated ingress", func() {
			cases := NewDefaultDiscreteGenerator(true, "1.2.3.4").GenerateTestCases()
			Expect(len(cases)).To(Equal(43))
		})
	})
}
