package connectivity

import (
	"fmt"
	connectivitykube "github.com/mattfenwick/cyclonus/pkg/connectivity/kube"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/synthetic"
	"github.com/mattfenwick/cyclonus/pkg/explainer"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/yaml"
	"time"
)

type Tester struct {
	kubernetes *kube.Kubernetes
}

func NewTester(kubernetes *kube.Kubernetes) *Tester {
	return &Tester{kubernetes: kubernetes}
}

type TestCase struct {
	KubePolicy                *networkingv1.NetworkPolicy
	Noisy                     bool // TODO move into tester property ?
	NetpolCreationWaitSeconds int
	Port                      int
	Protocol                  v1.Protocol
	KubeResources             *connectivitykube.Resources
	SyntheticResources        *synthetic.Resources
	IgnoreLoopback            bool
	NamespacesToClean         []string
}

func (t *Tester) TestNetworkPolicy(testCase *TestCase) *TestCaseResult {
	utils.DoOrDie(t.kubernetes.DeleteAllNetworkPoliciesInNamespaces(testCase.NamespacesToClean))

	policy := matcher.BuildNetworkPolicy(testCase.KubePolicy)

	result := &TestCaseResult{
		TestCase: testCase,
	}

	if testCase.Noisy {
		policyBytes, err := yaml.Marshal(testCase.KubePolicy)
		if err != nil {
			result.Err = errors.Wrapf(err, "unable to marshal network policy to yaml")
			return result
		}

		logrus.Infof("Creating network policy:\n%s\n\n", policyBytes)

		fmt.Printf("%s\n\n", explainer.Explain(policy))
		explainer.TableExplainer(policy).Render()
	}

	_, err := t.kubernetes.CreateNetworkPolicy(testCase.KubePolicy)
	if err != nil {
		result.Err = errors.Wrapf(err, "unable to create network policy")
		return result
	}

	logrus.Infof("waiting %d seconds for network policy to create and become active", testCase.NetpolCreationWaitSeconds)
	time.Sleep(time.Duration(testCase.NetpolCreationWaitSeconds) * time.Second)

	logrus.Infof("probe on port %d, protocol %s", testCase.Port, testCase.Protocol)

	result.SyntheticResult = synthetic.RunSyntheticProbe(&synthetic.Request{
		Protocol:  testCase.Protocol,
		Port:      testCase.Port,
		Policies:  policy,
		Resources: testCase.SyntheticResources,
	})

	result.KubeResult = connectivitykube.RunKubeProbe(t.kubernetes, &connectivitykube.Request{
		Resources:       testCase.KubeResources,
		Port:            testCase.Port,
		Protocol:        testCase.Protocol,
		NumberOfWorkers: 5,
	})

	fmt.Printf("\n\nKube results for %s/%s:\n", testCase.KubePolicy.Namespace, testCase.KubePolicy.Name)
	kubeProbe := result.KubeResult.TruthTable()
	kubeProbe.Table().Render()

	comparison := result.SyntheticResult.Combined.Compare(kubeProbe)
	trues, falses, nv, checked := comparison.ValueCounts(testCase.IgnoreLoopback)
	if falses > 0 {
		fmt.Printf("Discrepancy found: %d wrong, %d no value, %d correct out of %d total\n", falses, trues, nv, checked)
	} else {
		fmt.Printf("found %d true, %d false, %d no value from %d total\n", trues, falses, nv, checked)
	}

	if falses > 0 || testCase.Noisy {
		fmt.Println("Ingress:")
		result.SyntheticResult.Ingress.Table().Render()

		fmt.Println("Egress:")
		result.SyntheticResult.Egress.Table().Render()

		fmt.Println("Combined:")
		result.SyntheticResult.Combined.Table().Render()

		fmt.Printf("\n\nSynthetic vs combined:\n")
		comparison.Table().Render()
	}

	return result
}
