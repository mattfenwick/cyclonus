package connectivity

import (
	"encoding/xml"
	junit "github.com/jstemmer/go-junit-report/formatter"
	"github.com/sirupsen/logrus"
	"os"
)

type JUnitTestResult struct {
	Passed bool
	Name   string
}

func PrintJUnitResults(filename string, results []*Result, ignoreLoopback bool) error {
	if filename == "" {
		return nil
	}

	var junitResults []*JUnitTestResult
	for _, result := range results {
		junitResults = append(junitResults, &JUnitTestResult{
			Passed: result.Passed(ignoreLoopback),
			Name:   result.TestCase.Description,
		})
	}

	f, err := os.Create(filename)
	if err != nil {
		logrus.Errorf("Unable to create file %q for junit output: %v\n", filename, err)
		return err
	}
	defer f.Close()

	junitTestSuite := ResultsToJUnit(junitResults)
	enc := xml.NewEncoder(f)
	enc.Indent("", "    ")
	return enc.Encode(junitTestSuite)
}

func ResultsToJUnit(results []*JUnitTestResult) junit.JUnitTestSuite {
	var testCases []junit.JUnitTestCase
	failed := 0

	for _, result := range results {
		testCase := junit.JUnitTestCase{
			Name: result.Name,
		}
		if !result.Passed {
			testCase.Failure = &junit.JUnitFailure{}
			failed++
		}
		testCases = append(testCases, testCase)
	}
	return junit.JUnitTestSuite{
		Name:      "cyclonus",
		Failures:  failed,
		TestCases: testCases,
	}
}
