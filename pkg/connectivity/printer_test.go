package connectivity

import (
	"bytes"
	"flag"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
)

var update = flag.Bool("update", false, "update .golden files")

func RunPrinterTests() {
	Describe("Test junit file print", func() {
		testCases := []struct {
			desc       string
			results    []*JUnitTestResult
			expectFile string
			expectErr  string
		}{
			{
				desc:       "Empty summary",
				results:    nil,
				expectFile: "testdata/empty-summary.golden.xml",
			}, {
				desc: "2 pass 2 fail",
				results: []*JUnitTestResult{
					{Name: "test1", Passed: true},
					{Name: "test2 with spaces", Passed: false},
					{Name: "test3 with + special %chars/", Passed: true},
					{Name: "test4 with\nnewlines", Passed: false},
				},
				expectFile: "testdata/2-pass-2-fail.golden.xml",
			},
		}
		for _, testCase := range testCases {
			var output bytes.Buffer
			err := printJunit(&output, testCase.results)
			if err != nil {
				Expect(len(testCase.expectErr)).ToNot(Equal(0))
				Expect(err.Error()).To(Equal(testCase.expectErr))
			}

			if *update {
				err := ioutil.WriteFile(testCase.expectFile, output.Bytes(), 0666)
				Expect(err).To(BeNil())
			} else {
				fileData, err := ioutil.ReadFile(testCase.expectFile)
				Expect(err).To(BeNil())
				fmt.Printf("expected: \n%s\n", fileData)
				fmt.Printf("actual: \n%s\n", output.Bytes())
				Expect(bytes.Equal(fileData, output.Bytes())).To(BeTrue())
			}
		}
	})
}
