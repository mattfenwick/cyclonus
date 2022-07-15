package connectivity

import (
	junit "github.com/jstemmer/go-junit-report/formatter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func RunPrinterTests() {
	Describe("JUnit cyclonus output", func() {
		It("should convert results to junit", func() {
			testCases := []struct {
				desc    string
				results []*JUnitTestResult
				junit   junit.JUnitTestSuite
			}{
				{
					desc:    "Empty summary",
					results: nil,
					junit: junit.JUnitTestSuite{
						Failures:  0,
						Name:      "cyclonus",
						TestCases: nil,
					},
				}, {
					desc: "2 pass 2 fail",
					results: []*JUnitTestResult{
						{Name: "test1", Passed: true},
						{Name: "test2 with spaces", Passed: false},
						{Name: "test3 with + special %chars/", Passed: true},
						{Name: "test4 with\nnewlines", Passed: false},
					},
					junit: junit.JUnitTestSuite{
						Failures: 2,
						Name:     "cyclonus",
						TestCases: []junit.JUnitTestCase{
							{Name: "test1", Failure: nil},
							{Name: "test2 with spaces", Failure: &junit.JUnitFailure{}},
							{Name: "test3 with + special %chars/", Failure: nil},
							{Name: "test4 with\nnewlines", Failure: &junit.JUnitFailure{}},
						},
					},
				},
			}
			for _, testCase := range testCases {
				actual := ResultsToJUnit(testCase.results)
				Expect(actual).To(Equal(testCase.junit))
			}
		})
	})
}
