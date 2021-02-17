package connectivity

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/explainer"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math"
	"sigs.k8s.io/yaml"
	"sort"
	"strings"
)

type Printer struct {
	Noisy          bool
	IgnoreLoopback bool
	Results        []*Result
}

func (t *Printer) PrintSummary() {
	tableString := &strings.Builder{}
	tableString.WriteString("Summary:\n")
	table := tablewriter.NewWriter(tableString)
	table.SetRowLine(true)

	table.SetHeader([]string{"Test", "Result", "Step/Try", "Wrong", "Right", "Ignored", "TCP", "SCTP", "UDP"})

	passedTotal, failedTotal := 0, 0
	passFailCounts := map[bool]map[string]int{false: {}, true: {}}
	protocolCounts := map[v1.Protocol]map[Comparison]int{v1.ProtocolTCP: {}, v1.ProtocolSCTP: {}, v1.ProtocolUDP: {}}

	for testNumber, result := range t.Results {
		// preprocess to figure out whether it passed or failed
		passed := true
		for _, step := range result.Steps {
			if step.LastComparison().ValueCounts(t.IgnoreLoopback)[DifferentComparison] > 0 {
				passed = false
			}
		}

		for _, feature := range result.Features() {
			passFailCounts[passed][feature]++
		}

		var testResult string
		if passed {
			testResult = "passed"
			passedTotal++
		} else {
			testResult = "failed"
			failedTotal++
		}

		table.Append([]string{
			fmt.Sprintf("%d: %s", testNumber+1, result.TestCase.Description),
			testResult, "", "", "", "",
			"", "", "",
		})

		for stepNumber, step := range result.Steps {
			for tryNumber := range step.KubeProbes {
				counts := step.Comparison(tryNumber).ValueCounts(t.IgnoreLoopback)
				tryProtocolCounts := step.Comparison(tryNumber).ValueCountsByProtocol(t.IgnoreLoopback)
				tcp := tryProtocolCounts[v1.ProtocolTCP]
				sctp := tryProtocolCounts[v1.ProtocolSCTP]
				udp := tryProtocolCounts[v1.ProtocolUDP]
				table.Append([]string{
					"",
					"",
					fmt.Sprintf("Step %d, try %d", stepNumber+1, tryNumber+1),
					intToString(counts[DifferentComparison]),
					intToString(counts[SameComparison]),
					intToString(counts[IgnoredComparison]),
					protocolResult(tcp[SameComparison], tcp[DifferentComparison]),
					protocolResult(sctp[SameComparison], sctp[DifferentComparison]),
					protocolResult(udp[SameComparison], udp[DifferentComparison]),
				})

				protocolCounts[v1.ProtocolTCP][SameComparison] += tcp[SameComparison]
				protocolCounts[v1.ProtocolTCP][DifferentComparison] += tcp[DifferentComparison]
				protocolCounts[v1.ProtocolSCTP][SameComparison] += sctp[SameComparison]
				protocolCounts[v1.ProtocolSCTP][DifferentComparison] += sctp[DifferentComparison]
				protocolCounts[v1.ProtocolUDP][SameComparison] += udp[SameComparison]
				protocolCounts[v1.ProtocolUDP][DifferentComparison] += udp[DifferentComparison]
			}
		}
	}

	table.Render()
	fmt.Println(tableString.String())

	fmt.Println(passFailTable(passFailCounts, passedTotal, failedTotal, protocolCounts))
}

func protocolResult(passed int, failed int) string {
	total := passed + failed
	if total == 0 {
		return "-"
	}
	return fmt.Sprintf("%d / %d (%.0f%%)", passed, total, percentage(passed, total))
}

type passFailRow struct {
	Feature string
	Passed  int
	Failed  int
}

func (p *passFailRow) PassedPercentage() float64 {
	return percentage(p.Passed, p.Passed+p.Failed)
}

func passFailTable(passFailCounts map[bool]map[string]int, passedTotal int, failedTotal int, protocolCounts map[v1.Protocol]map[Comparison]int) string {
	str := &strings.Builder{}
	table := tablewriter.NewWriter(str)
	table.SetAutoWrapText(false)
	str.WriteString("Pass/Fail counts:\n")

	table.SetHeader([]string{"Feature", "Passed", "Failed", "Passed %"})

	allFeatures := map[string]bool{}
	for _, t := range []bool{false, true} {
		for f := range passFailCounts[t] {
			allFeatures[f] = true
		}
	}

	var rows []*passFailRow
	for feature := range allFeatures {
		rows = append(rows, &passFailRow{
			Feature: feature,
			Passed:  passFailCounts[true][feature],
			Failed:  passFailCounts[false][feature],
		})
	}
	for protocol, counts := range protocolCounts {
		rows = append(rows, &passFailRow{
			Feature: fmt.Sprintf("probe on %s", protocol),
			Passed:  counts[SameComparison],
			Failed:  counts[DifferentComparison],
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].PassedPercentage() < rows[j].PassedPercentage()
	})
	rows = append(rows, &passFailRow{Feature: "Total", Passed: passedTotal, Failed: failedTotal})

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
	return math.Round(100 * float64(i) / float64(total))
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
	fmt.Println("features:")
	for _, feature := range result.TestCase.GetFeatures().Strings() {
		fmt.Printf(" - %s\n", feature)
	}

	fmt.Printf("\n\n")
}

func (t *Printer) PrintStep(i int, step *generator.TestStep, stepResult *StepResult) {
	if step.Probe.PortProtocol != nil {
		fmt.Printf("step %d on port %s, protocol %s:\n", i, step.Probe.PortProtocol.Port.String(), step.Probe.PortProtocol.Protocol)
	} else {
		fmt.Printf("step %d on all available ports/protocols:\n", i)
	}
	policy := stepResult.Policy

	fmt.Printf("Policy explanation:\n%s\n", explainer.TableExplainer(policy))

	fmt.Printf("\n\nResults for network policies:\n")
	for _, netpol := range stepResult.KubePolicies {
		fmt.Printf(" - %s/%s:\n", netpol.Namespace, netpol.Name)
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

		if len(stepResult.KubePolicies) > 0 {
			for _, p := range stepResult.KubePolicies {
				fmt.Printf("Network policy:\n\n%s\n", PrintNetworkPolicy(p))
			}
		} else {
			fmt.Println("no network policies")
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
