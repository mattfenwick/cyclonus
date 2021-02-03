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
			lastKubeResult := step.LastKubeResult()
			counts := step.SyntheticResult.Combined.Compare(lastKubeResult.TruthTable()).ValueCounts(t.IgnoreLoopback)
			if counts.False > 0 {
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
			for tryNumber, kubeResult := range step.KubeResults {
				comparison := step.SyntheticResult.Combined.Compare(kubeResult.TruthTable())
				counts := comparison.ValueCounts(t.IgnoreLoopback)
				table.Append([]string{"", "", intToString(stepNumber + 1), intToString(tryNumber + 1), intToString(counts.False), intToString(counts.True)})
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
		fmt.Printf("test case failed to execute for %+v: %+v", result.TestCase, result.Err)
		return
	}

	stepCount := len(result.TestCase.Steps)
	resultCount := len(result.Steps)
	if stepCount != resultCount {
		panic(errors.Errorf("found %d test steps, but %d result steps", stepCount, resultCount))
	}

	for i := range result.Steps {
		t.PrintStep(i, result.TestCase.Steps[i], result.Steps[i])
	}

	fmt.Printf("\n\n")
}

func (t *Printer) PrintStep(i int, step *generator.TestStep, stepResult *StepResult) {
	fmt.Printf("step %d on port %s, protocol %s:\n", i, step.Port.String(), step.Protocol)
	policy := stepResult.Policy

	if t.Noisy {
		//fmt.Printf("%s\n\n", explainer.Explain(policy))
		explainer.TableExplainer(policy).Render()
	}

	fmt.Printf("\n\nKube results for:\n")
	for _, netpol := range stepResult.KubePolicies {
		fmt.Printf("  policy %s/%s:\n", netpol.Namespace, netpol.Name)
	}

	if len(stepResult.KubeResults) == 0 {
		panic(errors.Errorf("found 0 KubeResults for step, expected 1 or more"))
	}

	lastKubeProbe := stepResult.LastKubeResult().TruthTable()
	fmt.Println(lastKubeProbe.Table())

	comparison := stepResult.SyntheticResult.Combined.Compare(lastKubeProbe)
	counts := comparison.ValueCounts(t.IgnoreLoopback)
	if counts.False > 0 {
		fmt.Printf("Discrepancy found:")
	}
	fmt.Printf("%d wrong, %d no value, %d correct, %d ignored out of %d total\n", counts.False, counts.True, counts.NoValue, counts.Ignored, counts.Total)

	if counts.False > 0 || t.Noisy {
		//fmt.Println("Ingress:")
		//step.SyntheticResult.Ingress.Table().Render()
		//
		//fmt.Println("Egress:")
		//step.SyntheticResult.Egress.Table().Render()

		fmt.Println("Expected:")
		fmt.Println(stepResult.SyntheticResult.Combined.Table())

		for i, kubeResult := range stepResult.KubeResults {
			fmt.Printf("kube results, try %d:\n", i)
			fmt.Println(kubeResult.TruthTable().Table())
		}

		if len(stepResult.KubePolicies) > 0 {
			for _, p := range stepResult.KubePolicies {
				fmt.Printf("Network policy:\n\n%s\n", PrintNetworkPolicy(p))
			}
		} else {
			fmt.Println("no network policies")
		}

		fmt.Printf("\nActual vs expected (last round):\n")
		fmt.Println(comparison.Table())
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
