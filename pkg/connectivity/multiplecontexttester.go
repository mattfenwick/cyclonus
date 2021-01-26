package connectivity

import (
	"fmt"
	connectivitykube "github.com/mattfenwick/cyclonus/pkg/connectivity/kube"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/synthetic"
	"github.com/mattfenwick/cyclonus/pkg/explainer"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/yaml"
	"time"
)

type MultipleContextTestCase struct {
	KubePolicies              []*networkingv1.NetworkPolicy
	NetpolCreationWaitSeconds int
	Port                      int
	Protocol                  v1.Protocol
	KubeClients               map[string]*kube.Kubernetes
	KubeResources             map[string]*connectivitykube.Resources
	SyntheticResources        *synthetic.Resources
	NamespacesToClean         []string
	Policy                    *matcher.Policy
}

type MultipleContextTestCaseResult struct {
	TestCase        *MultipleContextTestCase
	SyntheticResult *synthetic.Result
	KubeResults     map[string]*connectivitykube.Results
	Errors          []error
}

type MultipleContextTester struct{}

func NewMultipleContextTester() *MultipleContextTester {
	return &MultipleContextTester{}
}

type workerResult struct {
	ContextName string
	Results     *connectivitykube.Results
	Err         error
}

func (t *MultipleContextTester) TestNetworkPolicy(testCase *MultipleContextTestCase) *MultipleContextTestCaseResult {
	result := &MultipleContextTestCaseResult{
		TestCase: testCase,
		SyntheticResult: synthetic.RunSyntheticProbe(&synthetic.Request{
			Protocol:  testCase.Protocol,
			Port:      testCase.Port,
			Resources: testCase.SyntheticResources,
			Policies:  testCase.Policy,
		}),
		KubeResults: map[string]*connectivitykube.Results{},
	}

	resultChan := make(chan *workerResult, len(testCase.KubeClients))
	for contextName := range testCase.KubeClients {
		go func(context string) {
			results, err := t.testWorker(testCase, context)
			resultChan <- &workerResult{
				ContextName: context,
				Results:     results,
				Err:         err,
			}
		}(contextName)
	}

	for i := 0; i < len(testCase.KubeClients); i++ {
		results := <-resultChan
		if results.Err == nil {
			result.KubeResults[results.ContextName] = results.Results
		} else {
			result.Errors = append(result.Errors, results.Err)
		}
	}

	return result
}

func (t *MultipleContextTester) testWorker(testCase *MultipleContextTestCase, contextName string) (*connectivitykube.Results, error) {
	kubeClient := testCase.KubeClients[contextName]
	utils.DoOrDie(kubeClient.DeleteAllNetworkPoliciesInNamespaces(testCase.NamespacesToClean))

	for _, kubePolicy := range testCase.KubePolicies {
		_, err := kubeClient.CreateNetworkPolicy(kubePolicy)
		if err != nil {
			return nil, err
		}
	}

	logrus.Infof("waiting %d seconds for network policy to create and become active", testCase.NetpolCreationWaitSeconds)
	time.Sleep(time.Duration(testCase.NetpolCreationWaitSeconds) * time.Second)

	logrus.Infof("running probe on context %s", contextName)

	return connectivitykube.RunKubeProbe(kubeClient, &connectivitykube.Request{
		Resources:       testCase.KubeResources[contextName],
		Port:            testCase.Port,
		Protocol:        testCase.Protocol,
		NumberOfWorkers: 5,
	}), nil
}

type MultipleContextTestCasePrinter struct {
	Noisy          bool
	IgnoreLoopback bool
}

func (t *MultipleContextTestCasePrinter) PrintTestCaseResult(result *MultipleContextTestCaseResult) {
	policy := result.TestCase.Policy

	policyBytes, err := yaml.Marshal(result.TestCase.KubePolicies)
	utils.DoOrDie(err)
	fmt.Printf("Network policy:\n\n%s\n", policyBytes)

	explainer.TableExplainer(policy).Render()

	//fmt.Println("expected results:")
	//fmt.Println("Expected ingress:")
	//result.SyntheticResult.Ingress.Table().Render()
	//
	//fmt.Println("Expected egress:")
	//result.SyntheticResult.Egress.Table().Render()

	fmt.Println("Expected connectivity:")
	result.SyntheticResult.Combined.Table().Render()

	foundDiscrepancy := false

	for contextName, kubeResults := range result.KubeResults {
		comparison := result.SyntheticResult.Combined.Compare(kubeResults.TruthTable())
		_, falses, _, _ := comparison.ValueCounts(t.IgnoreLoopback)
		if falses > 0 {
			fmt.Printf("found %d discrepancies from expected in context %s\n", falses, contextName)
			foundDiscrepancy = true
		}
	}

	if foundDiscrepancy {
		for context, kubeResults := range result.KubeResults {
			kubeProbe := kubeResults.TruthTable()
			comparison := result.SyntheticResult.Combined.Compare(kubeProbe)
			trues, falses, nv, checked := comparison.ValueCounts(t.IgnoreLoopback)

			fmt.Printf("results for context %s:\n", context)
			fmt.Printf("found %d true, %d false, %d no value from %d total\n", trues, falses, nv, checked)
			kubeProbe.Table().Render()
		}
	} else {
		fmt.Println("no differences found for policy")
	}
}
