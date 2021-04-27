package connectivity

import (
	"bytes"
	"flag"
	"io/ioutil"
	"testing"
)

var update = flag.Bool("update", false, "update .golden files")

func TestPrintJunit(t *testing.T) {
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
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var output bytes.Buffer
			err := printJunit(&output, tC.summary)
			if err != nil {
				if len(tC.expectErr) == 0 {
					t.Fatalf("Expected nil error but got %v", err)
				}
				if err.Error() != tC.expectErr {
					t.Fatalf("Expected error %q but got %q", err, tC.expectErr)
				}
			}

			if *update {
				ioutil.WriteFile(tC.expectFile, output.Bytes(), 0666)
			} else {
				fileData, err := ioutil.ReadFile(tC.expectFile)
				if err != nil {
					t.Fatalf("Failed to read golden file %v: %v", tC.expectFile, err)
				}
				if !bytes.Equal(fileData, output.Bytes()) {
					t.Errorf("Expected junit to equal goldenfile: %v but instead got:\n\n%v", tC.expectFile, output.String())
				}
			}
		})
	}
}
