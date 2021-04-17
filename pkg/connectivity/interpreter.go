package connectivity

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/probe"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	networkingv1 "k8s.io/api/networking/v1"
	"time"
)

const (
	defaultWorkersCount = 15

	// 9 = 3 namespaces x 3 pods
	defaultBatchWorkersCount = 9
)

type InterpreterConfig struct {
	ResetClusterBeforeTestCase       bool
	KubeProbeRetries                 int
	PerturbationWaitSeconds          int
	VerifyClusterStateBeforeTestCase bool
	BatchJobs                        bool
	IgnoreLoopback                   bool
	JobTimeoutSeconds                int
}

func (i *InterpreterConfig) PerturbationWaitDuration() time.Duration {
	return time.Duration(i.PerturbationWaitSeconds) * time.Second
}

type Interpreter struct {
	kubernetes kube.IKubernetes
	resources  *probe.Resources
	kubeRunner *probe.Runner
	jobBuilder *probe.JobBuilder
	Config     *InterpreterConfig
}

func NewInterpreter(kubernetes kube.IKubernetes, resources *probe.Resources, config *InterpreterConfig) *Interpreter {
	fmt.Printf("resources:\n%s\n", resources.RenderTable())

	jobBuilder := &probe.JobBuilder{TimeoutSeconds: config.JobTimeoutSeconds}
	var kubeRunner *probe.Runner
	if config.BatchJobs {
		kubeRunner = probe.NewKubeBatchRunner(kubernetes, defaultBatchWorkersCount, jobBuilder)
	} else {
		kubeRunner = probe.NewKubeRunner(kubernetes, defaultWorkersCount, jobBuilder)
	}

	return &Interpreter{
		kubernetes: kubernetes,
		resources:  resources,
		kubeRunner: kubeRunner,
		jobBuilder: jobBuilder,
		Config:     config,
	}
}

func (t *Interpreter) ExecuteTestCase(testCase *generator.TestCase) *Result {
	result := &Result{InitialResources: t.resources, TestCase: testCase}
	var err error

	// keep track of what's in the cluster, so that we can correctly simulate expected results
	testCaseState := &TestCaseState{
		Kubernetes: t.kubernetes,
		Resources:  t.resources,
		Policies:   []*networkingv1.NetworkPolicy{},
	}

	if t.Config.ResetClusterBeforeTestCase {
		err = testCaseState.ResetClusterState()
		if err != nil {
			result.Err = err
			return result
		}
		logrus.Info("cluster state reset")
	}

	if t.Config.VerifyClusterStateBeforeTestCase {
		err = testCaseState.VerifyClusterState()
		if err != nil {
			result.Err = err
			return result
		}
		logrus.Info("cluster state verified")
	}

	// perform perturbations one at a time, and run a probe after each change
	for stepIndex, step := range testCase.Steps {
		// TODO grab actual netpols from kube and record in results, for extra debugging/sanity checks

		for actionIndex, action := range step.Actions {
			if action.CreatePolicy != nil {
				err = testCaseState.CreatePolicy(action.CreatePolicy.Policy)
			} else if action.UpdatePolicy != nil {
				err = testCaseState.UpdatePolicy(action.UpdatePolicy.Policy)
			} else if action.DeletePolicy != nil {
				err = testCaseState.DeletePolicy(action.DeletePolicy.Namespace, action.DeletePolicy.Name)
			} else if action.CreateNamespace != nil {
				err = testCaseState.CreateNamespace(action.CreateNamespace.Namespace, action.CreateNamespace.Labels)
			} else if action.SetNamespaceLabels != nil {
				err = testCaseState.SetNamespaceLabels(action.SetNamespaceLabels.Namespace, action.SetNamespaceLabels.Labels)
			} else if action.DeleteNamespace != nil {
				err = testCaseState.DeleteNamespace(action.DeleteNamespace.Namespace)
			} else if action.ReadNetworkPolicies != nil {
				err = testCaseState.ReadPolicies(action.ReadNetworkPolicies.Namespaces)
			} else if action.CreatePod != nil {
				err = testCaseState.CreatePod(action.CreatePod.Namespace, action.CreatePod.Pod, action.CreatePod.Labels)
			} else if action.SetPodLabels != nil {
				ns, pod, labels := action.SetPodLabels.Namespace, action.SetPodLabels.Pod, action.SetPodLabels.Labels
				err = testCaseState.SetPodLabels(ns, pod, labels)
			} else if action.DeletePod != nil {
				err = testCaseState.DeletePod(action.DeletePod.Namespace, action.DeletePod.Pod)
			} else {
				err = errors.Errorf("invalid Action at step %d, action %d", stepIndex, actionIndex)
			}
			if err != nil {
				result.Err = err
				return result
			}
		}

		logrus.Infof("step %d: waiting %d seconds for perturbation to take effect", stepIndex+1, t.Config.PerturbationWaitSeconds)
		time.Sleep(t.Config.PerturbationWaitDuration())

		result.Steps = append(result.Steps, t.runProbe(testCaseState, step.Probe))
	}

	return result
}

func (t *Interpreter) runProbe(testCaseState *TestCaseState, probeConfig *generator.ProbeConfig) *StepResult {
	parsedPolicy := matcher.BuildNetworkPolicies(true, testCaseState.Policies)

	logrus.Infof("running probe %+v", probeConfig)
	logrus.Debugf("with resources:\n%s", testCaseState.Resources.RenderTable())

	simRunner := probe.NewSimulatedRunner(parsedPolicy, t.jobBuilder)

	stepResult := NewStepResult(
		simRunner.RunProbeForConfig(probeConfig, testCaseState.Resources),
		parsedPolicy,
		append([]*networkingv1.NetworkPolicy{}, testCaseState.Policies...)) // this looks weird, but just making a new copy to avoid accidentally mutating it elsewhere

	for i := 0; i <= t.Config.KubeProbeRetries; i++ {
		logrus.Infof("running kube probe on try %d", i+1)
		stepResult.AddKubeProbe(t.kubeRunner.RunProbeForConfig(probeConfig, testCaseState.Resources))
		// no differences between synthetic and kube probes?  then we can stop
		if stepResult.LastComparison().ValueCounts(t.Config.IgnoreLoopback)[DifferentComparison] == 0 {
			break
		}
	}

	return stepResult
}
