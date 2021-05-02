package connectivity

import (
	"bytes"
	"flag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

var update = flag.Bool("update", false, "update .golden files")

func RunPrinterTests() {
	Describe("Test junit file print", func() {
		testCases := []struct {
			desc       string
			summary    *Summary
			expectFile string
			expectErr  string
		}{
			{
				desc:       "Empty summary",
				summary:    &Summary{},
				expectFile: "testdata/empty-summary.golden.xml",
			}, {
				desc: "2 pass 2 fail",
				summary: &Summary{
					Tests: [][]string{
						{"1: test1", "passed", ""},
						{"2: test2 with spaces", "failed", ""},
						{"3: test3 with + special %chars/", "passed", ""},
						{"4: test4 with\nnewlines", "foo-is-failed", ""},
					},
				},
				expectFile: "testdata/2-pass-2-fail.golden.xml",
			}, {
				desc: "Uses failure count from summary",
				summary: &Summary{
					Failed: 10,
				},
				expectFile: "testdata/use-summary-failure-count.golden.xml",
			},
		}
		for _, testCase := range testCases {
			var output bytes.Buffer
			err := printJunit(&output, testCase.summary)
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
				logrus.Infof("expected: \n%s\n", fileData)
				logrus.Infof("actual: \n%s\n", output.Bytes())
				Expect(bytes.Equal(fileData, output.Bytes())).To(BeTrue())
			}
		}
	})
}
