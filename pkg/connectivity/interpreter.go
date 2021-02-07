package connectivity

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/types"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"time"
)

type Interpreter struct {
	kubernetes                       *kube.Kubernetes
	resources                        *types.Resources
	kubeProbeRetries                 int
	perturbationWaitDuration         time.Duration
	resetClusterBeforeTestCase       bool
	verifyClusterStateBeforeTestCase bool
}

func NewInterpreter(kubernetes *kube.Kubernetes, resources *types.Resources, resetClusterBeforeTestCase bool, kubeProbeRetries int, perturbationWaitSeconds int, verifyClusterStateBeforeTestCase bool) (*Interpreter, error) {
	fmt.Printf("resources:\n%s\n", resources.RenderTable())

	return &Interpreter{
		kubernetes:                       kubernetes,
		resources:                        resources,
		kubeProbeRetries:                 kubeProbeRetries,
		perturbationWaitDuration:         time.Duration(perturbationWaitSeconds) * time.Second,
		resetClusterBeforeTestCase:       resetClusterBeforeTestCase,
		verifyClusterStateBeforeTestCase: verifyClusterStateBeforeTestCase,
	}, nil
}

func (t *Interpreter) ExecuteTestCase(testCase *generator.TestCase) *Result {
	result := &Result{TestCase: testCase}
	var err error

	if t.resetClusterBeforeTestCase {
		err = t.resetClusterState()
		if err != nil {
			result.Err = err
			return result
		}
		logrus.Info("cluster state reset")
	}

	if t.verifyClusterStateBeforeTestCase {
		err = t.verifyClusterState()
		if err != nil {
			result.Err = err
			return result
		}
		logrus.Info("cluster state verified")
	}

	// keep track of what's in the cluster, so that we can correctly simulate expected results
	testCaseState := &TestCaseState{
		Kubernetes: t.kubernetes,
		Resources:  t.resources,
		Policies:   []*networkingv1.NetworkPolicy{},
	}

	// perform perturbations one at a time, and run a probe after each change
	for stepIndex, step := range testCase.Steps {
		// TODO grab actual netpols from kube and record in results, for extra debugging/sanity checks

		for actionIndex, action := range step.Actions {
			if action.CreatePolicy != nil {
				err = testCaseState.CreatePolicy(action.CreatePolicy.Policy)
			} else if action.UpdatePolicy != nil {
				err = testCaseState.UpdatePolicy(action.UpdatePolicy.Policy)
			} else if action.SetNamespaceLabels != nil {
				err = testCaseState.SetNamespaceLabels(action.SetNamespaceLabels.Namespace, action.SetNamespaceLabels.Labels)
			} else if action.SetPodLabels != nil {
				ns, pod, labels := action.SetPodLabels.Namespace, action.SetPodLabels.Pod, action.SetPodLabels.Labels
				err = testCaseState.SetPodLabels(ns, pod, labels)
			} else if action.ReadNetworkPolicies != nil {
				err = testCaseState.ReadPolicies(action.ReadNetworkPolicies.Namespaces)
			} else if action.DeletePolicy != nil {
				err = testCaseState.DeletePolicy(action.DeletePolicy.Namespace, action.DeletePolicy.Name)
			} else {
				err = errors.Errorf("invalid Action at step %d, action %d", stepIndex, actionIndex)
			}
			if err != nil {
				result.Err = err
				return result
			}
		}

		logrus.Infof("waiting %f seconds for perturbation to take effect", t.perturbationWaitDuration.Seconds())
		time.Sleep(t.perturbationWaitDuration)

		result.Steps = append(result.Steps, t.runProbe(testCaseState, step.Port, step.Protocol))
	}

	return result
}

func (t *Interpreter) runProbe(testCaseState *TestCaseState, port intstr.IntOrString, protocol v1.Protocol) *StepResult {
	parsedPolicy := matcher.BuildNetworkPolicies(testCaseState.Policies)

	logrus.Infof("running probe on port %s, protocol %s", port.String(), protocol)

	jobs := t.resources.GetJobsForSpecificPortProtocol(port, protocol)
	kubeRunner := types.NewKubeProbeRunner(t.kubernetes, 5)
	newTable := func() *types.Table { return t.resources.NewTable() }

	simRunner := types.NewSimulatedProbeRunner(parsedPolicy)

	stepResult := &StepResult{
		SimulatedProbe: simRunner.RunProbe(jobs, newTable),
		Policy:         parsedPolicy,
		KubePolicies:   append([]*networkingv1.NetworkPolicy{}, testCaseState.Policies...), // this looks weird, but just making a new copy to avoid accidentally mutating it elsewhere
	}

	for i := 0; i <= t.kubeProbeRetries; i++ {
		logrus.Infof("running kube probe on try %d", i)
		kubeProbe := kubeRunner.RunProbe(jobs, newTable)
		resultTable := NewResultTableFrom(kubeProbe.Combined, stepResult.SimulatedProbe.Combined)
		stepResult.KubeProbes = append(stepResult.KubeProbes, kubeProbe.Combined)
		// no differences between synthetic and kube probes?  then we can stop
		if resultTable.ValueCounts(false)[DifferentComparison] == 0 {
			break
		}
	}

	return stepResult
}

func (t *Interpreter) resetClusterState() error {
	err := t.kubernetes.DeleteAllNetworkPoliciesInNamespaces(t.resources.NamespacesSlice())
	if err != nil {
		return err
	}

	return t.resources.ResetLabelsInKube(t.kubernetes)
}

func (t *Interpreter) verifyClusterState() error {
	err := t.resources.VerifyClusterState(t.kubernetes)
	if err != nil {
		return err
	}

	policies, err := t.kubernetes.GetNetworkPoliciesInNamespaces(t.resources.NamespacesSlice())
	if err != nil {
		return err
	}
	if len(policies) > 0 {
		return errors.Errorf("expected 0 policies in namespaces %+v, found %d", t.resources.NamespacesSlice(), len(policies))
	}
	return nil
}
