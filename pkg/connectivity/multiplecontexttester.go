package connectivity

/*
TODO
import (
	"fmt"
	connectivitykube "github.com/mattfenwick/cyclonus/pkg/connectivity/kube"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/synthetic"
	"github.com/mattfenwick/cyclonus/pkg/explainer"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"os"
	"sigs.k8s.io/yaml"
	"strings"
	"time"
)

type MultipleContextTestCase struct {
	KubePolicies              []*networkingv1.NetworkPolicy
	NetpolCreationWaitSeconds int
	Port                      intstr.IntOrString
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
	Noisy            bool
	IgnoreLoopback   bool
	Contexts         []string
	DifferenceCounts [][]int
}

func (t *MultipleContextTestCasePrinter) PrintTestCaseResult(result *MultipleContextTestCaseResult) {
	policy := result.TestCase.Policy

	policyBytes, err := yaml.Marshal(result.TestCase.KubePolicies)
	utils.DoOrDie(err)
	fmt.Printf("Network policy:\n\n%s\n", policyBytes)

	fmt.Printf("Policy explained:\n%s\n", explainer.TableExplainer(policy))

	//fmt.Println("expected results:")
	//fmt.Println("Expected ingress:")
	//result.SyntheticResult.Ingress.Table().Render()
	//
	//fmt.Println("Expected egress:")
	//result.SyntheticResult.Egress.Table().Render()

	fmt.Println("Expected connectivity:")
	fmt.Println(result.SyntheticResult.Combined.Table())

	foundDiscrepancy := false

	var falseCounts []int
	for _, contextName := range t.Contexts {
		kubeResults := result.KubeResults[contextName]
		comparison := result.SyntheticResult.Combined.Compare(kubeResults.TruthTable())
		counts := comparison.ValueCounts(t.IgnoreLoopback)
		if counts.False > 0 {
			fmt.Printf("results for context %s:\n", contextName)
			fmt.Printf("found %d true, %d false, %d no value, %d ignored from %d total\n", counts.True, counts.False, counts.NoValue, counts.Ignored, counts.Total)
			foundDiscrepancy = true
		}
		falseCounts = append(falseCounts, counts.False)
	}
	t.DifferenceCounts = append(t.DifferenceCounts, falseCounts)

	if foundDiscrepancy {
		for _, contextName := range t.Contexts {
			fmt.Printf("results for context %s:\n", contextName)
			kubeResults := result.KubeResults[contextName]
			fmt.Println(kubeResults.TruthTable().Table())
		}
	} else {
		fmt.Println("no differences found for policy")
	}
}

func (t *MultipleContextTestCasePrinter) PrintFinish() {
	csv := []string{strings.Join(t.Contexts, ",")}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(t.Contexts)
	for _, falseCounts := range t.DifferenceCounts {
		var row []string
		for _, falses := range falseCounts {
			row = append(row, fmt.Sprintf("%d", falses))
		}
		csv = append(csv, strings.Join(row, ","))
		table.Append(row)
	}
	table.Render()

	fmt.Printf("\n\n%s\n\n", strings.Join(csv, "\n"))
}
*/

//func SetupClusterTODODelete(kubernetes *kube.Kubernetes, namespaces []string, pods []string, port int, protocol v1.Protocol) (*connectivitykube.Resources, *synthetic.Resources, error) {
//	kubeResources := connectivitykube.NewDefaultResources(namespaces, pods, []int{port}, []v1.Protocol{protocol})
//
//	err := kubeResources.CreateResourcesInKube(kubernetes)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	err = waitForPodsReadyTODODelete(kubernetes, namespaces, pods, 60)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	podList, err := kubernetes.GetPodsInNamespaces(namespaces)
//	if err != nil {
//		return nil, nil, err
//	}
//	var syntheticPods []*synthetic.Pod
//	for _, pod := range podList {
//		ip := pod.Status.PodIP
//		if ip == "" {
//			return nil, nil, errors.Errorf("no ip found for pod %s/%s", pod.Namespace, pod.Name)
//		}
//		syntheticPods = append(syntheticPods, &synthetic.Pod{
//			Namespace: pod.Namespace,
//			Name:      pod.Name,
//			Labels:    pod.Labels,
//			IP:        ip,
//		})
//		log.Infof("ip for pod %s/%s: %s", pod.Namespace, pod.Name, ip)
//	}
//
//	resources, err := synthetic.NewResources(kubeResources.Namespaces, syntheticPods)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	return kubeResources, resources, nil
//}

//func waitForPodsReadyTODODelete(kubernetes *kube.Kubernetes, namespaces []string, pods []string, timeoutSeconds int) error {
//	sleep := 5
//	for i := 0; i < timeoutSeconds; i += sleep {
//		podList, err := kubernetes.GetPodsInNamespaces(namespaces)
//		if err != nil {
//			return err
//		}
//
//		ready := 0
//		for _, pod := range podList {
//			if pod.Status.Phase == "Running" && pod.Status.PodIP != "" {
//				ready++
//			}
//		}
//		if ready == len(namespaces)*len(pods) {
//			return nil
//		}
//
//		log.Infof("waiting for pods to be running and have IP addresses")
//		time.Sleep(time.Duration(sleep) * time.Second)
//	}
//	return errors.Errorf("pods not ready")
//}
