package connectivity

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/explainer"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"sigs.k8s.io/yaml"
)

type TestCasePrinter struct {
	Noisy          bool
	IgnoreLoopback bool
}

func (t *TestCasePrinter) PrintTestCaseResult(result *TestCaseResult) {
	policy := result.Policy

	if t.Noisy {
		//fmt.Printf("%s\n\n", explainer.Explain(policy))
		explainer.TableExplainer(policy).Render()
	}

	fmt.Printf("\n\nKube results for %s/%s:\n", result.TestCase.KubePolicy.Namespace, result.TestCase.KubePolicy.Name)
	kubeProbe := result.KubeResult.TruthTable()
	kubeProbe.Table().Render()

	comparison := result.SyntheticResult.Combined.Compare(kubeProbe)
	trues, falses, nv, checked := comparison.ValueCounts(t.IgnoreLoopback)
	if falses > 0 {
		fmt.Printf("Discrepancy found: %d wrong, %d no value, %d correct out of %d total\n", falses, trues, nv, checked)
	} else {
		fmt.Printf("found %d true, %d false, %d no value from %d total\n", trues, falses, nv, checked)
	}

	if falses > 0 || t.Noisy {
		fmt.Println("Ingress:")
		result.SyntheticResult.Ingress.Table().Render()

		fmt.Println("Egress:")
		result.SyntheticResult.Egress.Table().Render()

		fmt.Println("Combined:")
		result.SyntheticResult.Combined.Table().Render()

		policyBytes, err := yaml.Marshal(result.TestCase.KubePolicy)
		utils.DoOrDie(err)
		fmt.Printf("Network policy:\n\n%s\n", policyBytes)

		fmt.Printf("\nSynthetic vs combined:\n")
		comparison.Table().Render()
	}
}
