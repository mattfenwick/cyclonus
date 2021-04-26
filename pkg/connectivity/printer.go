package connectivity

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	junit "github.com/jstemmer/go-junit-report/formatter"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type Printer struct {
	Noisy            bool
	IgnoreLoopback   bool
	JunitResultsFile string
	Results          []*Result
}

func (t *Printer) PrintSummary() {
	summary := (&CombinedResults{Results: t.Results}).Summary(t.IgnoreLoopback)
	t.printTestSummary(summary.Tests)
	for primary, counts := range summary.TagCounts {
		fmt.Println(passFailTable(primary, counts, nil, nil))
	}
	fmt.Println(protocolPassFailTable(summary.ProtocolCounts))

	fmt.Printf("Feature results:\n%s\n\n", t.printMarkdownFeatureTable(summary.FeaturePrimaryCounts, summary.FeatureCounts))
	fmt.Printf("Tag results:\n%s\n", t.printMarkdownFeatureTable(summary.TagPrimaryCounts, summary.TagCounts))
	if t.JunitResultsFile != "" {
		f, err := os.Create(t.JunitResultsFile)
		if err != nil {
			logrus.Errorf("Unable to create file %q for junit output: %v\n", t.JunitResultsFile, err)
		} else {
			defer f.Close()
			if err := printJunit(f, summary); err != nil {
				logrus.Errorf("Unable to write junit output: %v\n", err)
			}
		}
	}
}

func printJunit(w io.Writer, summary *Summary) error {
	s := summaryToJunit(summary)
	enc := xml.NewEncoder(w)
	return enc.Encode(s)
}

func summaryToJunit(summary *Summary) junit.JUnitTestSuite {
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

const (
	passSymbol = "\u2705"
	failSymbol = "\u274c"
)

type markdownRow struct {
	Name      string
	IsPrimary bool
	Pass      int
	Fail      int
}

func (m *markdownRow) GetName() string {
	if m.IsPrimary {
		return m.Name
	}
	return " - " + m.Name
}

func (m *markdownRow) symbol() string {
	if m.Fail == 0 {
		return passSymbol
	}
	return failSymbol
}

func (m *markdownRow) GetResult() string {
	total := m.Pass + m.Fail
	return fmt.Sprintf("%d / %d = %.0f%% %s", m.Pass, total, percentage(m.Pass, total), m.symbol())
}

func (t *Printer) printMarkdownFeatureTable(primaryCounts map[string]map[bool]int, tagCounts map[string]map[string]map[bool]int) string {
	var primaries []string
	for primary := range tagCounts {
		primaries = append(primaries, primary)
	}
	sort.Strings(primaries)

	var rows []*markdownRow
	for _, primary := range primaries {
		var subs []string
		for sub := range tagCounts[primary] {
			subs = append(subs, sub)
		}
		sort.Strings(subs)
		rows = append(rows, &markdownRow{Name: primary, IsPrimary: true, Pass: primaryCounts[primary][true], Fail: primaryCounts[primary][false]})
		for _, sub := range subs {
			counts := tagCounts[primary][sub]
			rows = append(rows, &markdownRow{
				Name:      sub,
				IsPrimary: false,
				Pass:      counts[true],
				Fail:      counts[false],
			})
		}
	}

	lines := []string{"| Tag | Result |", "| --- | --- |"}
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf("| %s | %s |", row.GetName(), row.GetResult()))
	}

	return strings.Join(lines, "\n")
}

func (t *Printer) printTestSummary(rows [][]string) {
	tableString := &strings.Builder{}
	tableString.WriteString("Summary:\n")
	table := tablewriter.NewWriter(tableString)
	table.SetRowLine(true)

	table.SetHeader([]string{"Test", "Result", "Step/Try", "Wrong", "Right", "Ignored", "TCP", "SCTP", "UDP"})

	table.AppendBulk(rows)

	table.Render()
	fmt.Println(tableString.String())
}

type passFailRow struct {
	Feature string
	Passed  int
	Failed  int
}

func (p *passFailRow) PassedPercentage() float64 {
	return percentage(p.Passed, p.Passed+p.Failed)
}

func passFailTable(caption string, passFailCounts map[string]map[bool]int, passedTotal *int, failedTotal *int) string {
	str := &strings.Builder{}
	table := tablewriter.NewWriter(str)
	table.SetAutoWrapText(false)
	str.WriteString(fmt.Sprintf("%s counts:\n", caption))

	table.SetHeader([]string{"Feature", "Passed", "Failed", "Passed %"})

	var rows []*passFailRow
	for feature := range passFailCounts {
		rows = append(rows, &passFailRow{
			Feature: feature,
			Passed:  passFailCounts[feature][true],
			Failed:  passFailCounts[feature][false],
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].PassedPercentage() < rows[j].PassedPercentage()
	})
	if passedTotal != nil || failedTotal != nil {
		rows = append(rows, &passFailRow{Feature: "Total", Passed: *passedTotal, Failed: *failedTotal})
	}

	for _, row := range rows {
		table.Append([]string{row.Feature, intToString(row.Passed), intToString(row.Failed), fmt.Sprintf("%.0f", row.PassedPercentage())})
	}

	table.Render()
	return str.String()
}

func protocolPassFailTable(protocolCounts map[v1.Protocol]map[Comparison]int) string {
	str := &strings.Builder{}
	table := tablewriter.NewWriter(str)
	table.SetAutoWrapText(false)
	str.WriteString("Pass/Fail for probes on protocols:\n")

	table.SetHeader([]string{"Protocol", "Passed", "Failed", "Passed %"})

	var rows []*passFailRow
	for protocol, counts := range protocolCounts {
		rows = append(rows, &passFailRow{
			Feature: fmt.Sprintf("probe on %s", protocol),
			Passed:  counts[SameComparison],
			Failed:  counts[DifferentComparison],
		})
	}

	for _, row := range rows {
		table.Append([]string{row.Feature, intToString(row.Passed), intToString(row.Failed), fmt.Sprintf("%.0f", row.PassedPercentage())})
	}

	table.Render()
	return str.String()
}

func percentage(i int, total int) float64 {
	if i+total == 0 {
		return 0
	}
	return math.Floor(100 * float64(i) / float64(total))
}

func intToString(i int) string {
	return fmt.Sprintf("%d", i)
}

func (t *Printer) PrintTestCaseResult(result *Result) {
	t.Results = append(t.Results, result)

	if result.Err != nil {
		fmt.Printf("test case failed to execute for %s %+v: %+v", result.TestCase.Description, result.TestCase, result.Err)
		return
	}

	fmt.Printf("evaluating test case: %s\n", result.TestCase.Description)
	stepCount := len(result.TestCase.Steps)
	resultCount := len(result.Steps)
	if stepCount != resultCount {
		panic(errors.Errorf("found %d test steps, but %d result steps", stepCount, resultCount))
	}

	for i := range result.Steps {
		t.PrintStep(i+1, result.TestCase.Steps[i], result.Steps[i])
	}
	//fmt.Println("features:")
	//for feature := range result.TestCase.GetFeatures() {
	//	fmt.Printf(" - %s\n", feature)
	//}

	fmt.Printf("\n\n")
}

func (t *Printer) PrintStep(i int, step *generator.TestStep, stepResult *StepResult) {
	if step.Probe.PortProtocol != nil {
		fmt.Printf("step %d on port %s, protocol %s:\n", i, step.Probe.PortProtocol.Port.String(), step.Probe.PortProtocol.Protocol)
	} else {
		fmt.Printf("step %d on all available ports/protocols:\n", i)
	}
	policy := stepResult.Policy

	fmt.Printf("Policy explanation:\n%s\n", policy.ExplainTable())

	fmt.Printf("\n\nResults for network policies:\n")
	if len(stepResult.KubePolicies) > 0 {
		for _, p := range stepResult.KubePolicies {
			fmt.Printf("Network policy:\n\n%s\n", PrintNetworkPolicy(p))
		}
	} else {
		fmt.Println("no network policies")
	}

	if len(stepResult.KubeProbes) == 0 {
		panic(errors.Errorf("found 0 KubeResults for step, expected 1 or more"))
	}

	comparison := stepResult.LastComparison()
	counts := comparison.ValueCounts(t.IgnoreLoopback)
	if counts[DifferentComparison] > 0 {
		fmt.Printf("Discrepancy found:")
	}
	fmt.Printf("%d wrong, %d ignored, %d correct\n", counts[DifferentComparison], counts[IgnoredComparison], counts[SameComparison])

	if counts[DifferentComparison] > 0 || t.Noisy {
		fmt.Printf("Expected ingress:\n%s\n", stepResult.SimulatedProbe.RenderIngress())

		fmt.Printf("Expected egress:\n%s\n", stepResult.SimulatedProbe.RenderEgress())

		fmt.Printf("Expected combined:\n%s\n", stepResult.SimulatedProbe.RenderTable())

		for i, kubeResult := range stepResult.KubeProbes {
			fmt.Printf("kube results, try %d:\n%s\n", i, kubeResult.RenderTable())
		}

		fmt.Printf("\nActual vs expected (last round):\n%s\n", comparison.RenderSuccessTable())
	} else {
		fmt.Printf("%s\n", stepResult.LastKubeProbe().RenderTable())
	}
}

func PrintNetworkPolicy(p *networkingv1.NetworkPolicy) string {
	// TODO is this a bad idea?
	// nil these out so the output isn't full of junk
	p.ManagedFields = nil
	p.UID = ""
	p.SelfLink = ""
	p.ResourceVersion = ""
	p.CreationTimestamp = metav1.Time{}
	p.Generation = 0

	policyBytes, err := yaml.Marshal(p)
	utils.DoOrDie(err)
	return string(policyBytes)
}
