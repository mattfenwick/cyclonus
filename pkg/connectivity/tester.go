package connectivity

import (
	connectivitykube "github.com/mattfenwick/cyclonus/pkg/connectivity/kube"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/synthetic"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type Tester struct {
	kubernetes *kube.Kubernetes
}

func NewTester(kubernetes *kube.Kubernetes) *Tester {
	return &Tester{kubernetes: kubernetes}
}

func (t *Tester) TestNetworkPolicy(testCase *TestCase) *TestCaseResult {
	utils.DoOrDie(t.kubernetes.DeleteAllNetworkPoliciesInNamespaces(testCase.NamespacesToClean))

	policy := matcher.BuildNetworkPolicy(testCase.KubePolicy)

	result := &TestCaseResult{
		TestCase: testCase,
		Policy:   policy,
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

	return result
}
