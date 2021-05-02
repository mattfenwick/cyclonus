package connectivity

import (
	"encoding/xml"
	junit "github.com/jstemmer/go-junit-report/formatter"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

type JUnitTestResult struct {
	Passed bool
	Name   string
}

func PrintJUnitResults(filename string, results []*Result, ignoreLoopback bool) {
	if filename == "" {
		return
	}

	f, err := os.Create(filename)
	if err != nil {
		logrus.Errorf("Unable to create file %q for junit output: %v\n", filename, err)
		return
	}

	var junitResults []*JUnitTestResult
	for _, result := range results {
		junitResults = append(junitResults, &JUnitTestResult{
			Passed: result.Passed(ignoreLoopback),
			Name:   result.TestCase.Description,
		})
	}

	defer f.Close()
	if err := printJunit(f, junitResults); err != nil {
		logrus.Errorf("Unable to write junit output: %v\n", err)
	}
}

func printJunit(w io.Writer, results []*JUnitTestResult) error {
	s := resultsToJUnit(results)
	enc := xml.NewEncoder(w)
	enc.Indent("", "    ")
	return enc.Encode(s)
}

func resultsToJUnit(results []*JUnitTestResult) junit.JUnitTestSuite {
	var testCases []junit.JUnitTestCase
	failed := 0

	for _, result := range results {
		testCase := junit.JUnitTestCase{
			Name: result.Name,
		}
		if !result.Passed {
			testCase.Failure = &junit.JUnitFailure{}
		}
		testCases = append(testCases, testCase)
	}
	return junit.JUnitTestSuite{
		Name:      "cyclonus",
		Failures:  failed,
		TestCases: testCases,
	}
}
