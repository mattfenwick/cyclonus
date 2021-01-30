package connectivity

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/explainer"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type Printer struct {
	Noisy          bool
	IgnoreLoopback bool
	Results        []*Result
}

func (t *Printer) PrintSummary() {
	fmt.Println("Summary:")
	for i, result := range t.Results {
		fmt.Printf("  test %d: %s\n", i, result.TestCase.Description)
		for j, step := range result.Steps {
			fmt.Printf("    step %d\n", j)
			for k, kubeResult := range step.KubeResults {
				fmt.Printf("      try %d\n", k)
				comparison := step.SyntheticResult.Combined.Compare(kubeResult.TruthTable())
				trues, falses, nv, checked := comparison.ValueCounts(t.IgnoreLoopback)
				fmt.Printf("        %d\t%d\t%d\t%d\n", trues, falses, nv, checked)
			}
		}
	}
	fmt.Println()
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
	fmt.Printf("step %d on port %d, protocol %s:\n", i, step.Port, step.Protocol)
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
	lastKubeProbe.Table().Render()

	comparison := stepResult.SyntheticResult.Combined.Compare(lastKubeProbe)
	trues, falses, nv, checked := comparison.ValueCounts(t.IgnoreLoopback)
	if falses > 0 {
		fmt.Printf("Discrepancy found: %d wrong, %d no value, %d correct out of %d total\n", falses, trues, nv, checked)
	} else {
		fmt.Printf("found %d true, %d false, %d no value from %d total\n", trues, falses, nv, checked)
	}

	if falses > 0 || t.Noisy {
		//fmt.Println("Ingress:")
		//step.SyntheticResult.Ingress.Table().Render()
		//
		//fmt.Println("Egress:")
		//step.SyntheticResult.Egress.Table().Render()

		fmt.Println("Expected:")
		stepResult.SyntheticResult.Combined.Table().Render()

		for i, kubeResult := range stepResult.KubeResults {
			fmt.Printf("kube results, try %d:\n", i)
			kubeResult.TruthTable().Table().Render()
		}

		if len(stepResult.KubePolicies) > 0 {
			for _, p := range stepResult.KubePolicies {
				fmt.Printf("Network policy:\n\n%s\n", PrintNetworkPolicy(p))
			}
		} else {
			fmt.Println("no network policies")
		}

		fmt.Printf("\nActual vs expected (last round):\n")
		comparison.Table().Render()
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
