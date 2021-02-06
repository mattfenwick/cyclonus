package connectivity

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/explainer"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
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

	table.SetHeader([]string{"Test", "Result", "Step", "Try", "Wrong", "Right"})

	for testNumber, result := range t.Results {
		// preprocess to figure out whether it passed or failed
		passed := true
		for _, step := range result.Steps {
			lastKubeResult := step.LastKubeProbe()
			comparison := NewResultTableFrom(lastKubeResult, step.SimulatedProbe.Combined)
			if comparison.ValueCounts(t.IgnoreLoopback)[DifferentComparison] > 0 {
				passed = false
			}
		}

		testResult := "success"
		if !passed {
			testResult = "failure"
		}
		table.Append([]string{
			fmt.Sprintf("%d: %s", testNumber+1, result.TestCase.Description),
			testResult, "", "", "", "",
		})

		for stepNumber, step := range result.Steps {
			for tryNumber, kubeProbe := range step.KubeProbes {
				comparison := NewResultTableFrom(kubeProbe, step.SimulatedProbe.Combined)
				counts := comparison.ValueCounts(t.IgnoreLoopback)
				table.Append([]string{"", "", intToString(stepNumber + 1), intToString(tryNumber + 1), intToString(counts[DifferentComparison]), intToString(counts[SameComparison])})
			}
		}
	}

	table.Render()
	fmt.Println(tableString.String())
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

	fmt.Printf("\n\n")
}

func (t *Printer) PrintStep(i int, step *generator.TestStep, stepResult *StepResult) {
	fmt.Printf("step %d on port %s, protocol %s:\n", i, step.Port.String(), step.Protocol)
	policy := stepResult.Policy

	if t.Noisy {
		fmt.Printf("Policy explained:\n%s\n", explainer.TableExplainer(policy))
	}

	fmt.Printf("\n\nKube results for:\n")
	for _, netpol := range stepResult.KubePolicies {
		fmt.Printf("  policy %s/%s:\n", netpol.Namespace, netpol.Name)
	}

	if len(stepResult.KubeProbes) == 0 {
		panic(errors.Errorf("found 0 KubeResults for step, expected 1 or more"))
	}

	lastKubeProbe := stepResult.LastKubeProbe()

	comparison := NewResultTableFrom(lastKubeProbe, stepResult.SimulatedProbe.Combined)
	counts := comparison.ValueCounts(t.IgnoreLoopback)
	if counts[DifferentComparison] > 0 {
		fmt.Printf("Discrepancy found:")
	}
	fmt.Printf("%d wrong, %d ignore, %d correct\n", counts[DifferentComparison], counts[IgnoredComparison], counts[SameComparison])

	if counts[DifferentComparison] > 0 || t.Noisy {
		fmt.Printf("Expected ingress:\n%s\n", stepResult.SimulatedProbe.Ingress.RenderTable())

		fmt.Printf("Expected egress:\n%s\n", stepResult.SimulatedProbe.Egress.RenderTable())

		fmt.Printf("Expected combined:\n%s\n", stepResult.SimulatedProbe.Combined.RenderTable())

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

		fmt.Printf("\nActual vs expected (last round):\n%s\n", comparison.RenderTable())
	} else {
		fmt.Printf("%s\n", lastKubeProbe.RenderTable())
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
