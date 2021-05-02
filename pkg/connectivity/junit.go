package connectivity

import (
	"encoding/xml"
	junit "github.com/jstemmer/go-junit-report/formatter"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strconv"
	"strings"
)

func PrintJUnitResults(filename string, summary *SummaryTable) {
	if filename == "" {
		return
	}

	f, err := os.Create(filename)
	if err != nil {
		logrus.Errorf("Unable to create file %q for junit output: %v\n", filename, err)
		return
	}

	defer f.Close()
	if err := printJunit(f, summary); err != nil {
		logrus.Errorf("Unable to write junit output: %v\n", err)
	}
}

func printJunit(w io.Writer, summary *SummaryTable) error {
	s := summaryToJunit(summary)
	enc := xml.NewEncoder(w)
	enc.Indent("", "    ")
	return enc.Encode(s)
}

func summaryToJunit(summary *SummaryTable) junit.JUnitTestSuite {
	s := junit.JUnitTestSuite{
		Name:      "cyclonus",
		Failures:  summary.Failed,
		TestCases: []junit.JUnitTestCase{},
	}

	for _, testStrings := range summary.Tests {
		_, testName, passed := parseTestStrings(testStrings)
		// Only cases where the testname are non-empty are new tests, otherwise it
		// is multi-line output of the test.
		if testName == "" {
			continue
		}
		tc := junit.JUnitTestCase{
			Name: testName,
		}
		if !passed {
			tc.Failure = &junit.JUnitFailure{}
		}
		s.TestCases = append(s.TestCases, tc)
	}
	return s
}

func parseTestStrings(input []string) (testNumber int, testName string, passed bool) {
	split := strings.SplitN(input[0], ": ", 2)
	if len(split) < 2 {
		return 0, "", false
	}

	testNumber, err := strconv.Atoi(split[0])
	if err != nil {
		logrus.Errorf("error parsing test number from string %q for junit: %v", split[0], err)
	}

	if len(input) > 1 && input[1] == "passed" {
		passed = true
	}

	return testNumber, split[1], passed
}
