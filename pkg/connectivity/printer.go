package connectivity

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/explainer"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type Printer struct {
	Noisy          bool
	IgnoreLoopback bool
}

func (t *Printer) PrintSummary() {
	fmt.Printf("TODO -- summary of number of passes, failures\n")
}

func (t *Printer) PrintTestCaseResult(result *Result) {
	if result.Err != nil {
		fmt.Printf("test case failed for %+v: %+v", result.TestCase, result.Err)
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

	kubeProbe := stepResult.KubeResult.TruthTable()
	kubeProbe.Table().Render()

	comparison := stepResult.SyntheticResult.Combined.Compare(kubeProbe)
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

		if len(stepResult.KubePolicies) > 0 {
			// TODO is this a bad idea?
			// nil these out so the output isn't full of junk
			for _, p := range stepResult.KubePolicies {
				p.ManagedFields = nil
				p.UID = ""
				p.SelfLink = ""
				p.ResourceVersion = ""
				p.CreationTimestamp = metav1.Time{}
				p.Generation = 0
			}

			policyBytes, err := yaml.Marshal(stepResult.KubePolicies)
			utils.DoOrDie(err)
			fmt.Printf("Network policy:\n\n%s\n", policyBytes)
		} else {
			fmt.Println("no network policies")
		}

		fmt.Printf("\nActual vs expected:\n")
		comparison.Table().Render()
	}
}
