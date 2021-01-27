package connectivity

import (
	connectivitykube "github.com/mattfenwick/cyclonus/pkg/connectivity/kube"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/synthetic"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/sirupsen/logrus"
	"time"
)

type Tester struct {
	kubernetes *kube.Kubernetes
	namespaces []string
}

func NewTester(kubernetes *kube.Kubernetes, namespaces []string) *Tester {
	return &Tester{kubernetes: kubernetes, namespaces: namespaces}
}

func (t *Tester) TestNetworkPolicy(testCase *TestCase) *TestCaseResult {
	utils.DoOrDie(t.kubernetes.DeleteAllNetworkPoliciesInNamespaces(t.namespaces))

	policy := matcher.BuildNetworkPolicies(testCase.KubePolicies)

	result := &TestCaseResult{
		TestCase: testCase,
		Policy:   policy,
	}

	for _, kubePolicy := range testCase.KubePolicies {
		_, err := t.kubernetes.CreateNetworkPolicy(kubePolicy)
		if err != nil {
			result.Err = err
			return result
		}
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

	return result
}
