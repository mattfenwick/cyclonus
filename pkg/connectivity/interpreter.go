package connectivity

import (
	connectivitykube "github.com/mattfenwick/cyclonus/pkg/connectivity/kube"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/synthetic"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"time"
)

type Interpreter struct {
	kubernetes                   *kube.Kubernetes
	kubeResources                *connectivitykube.Resources
	syntheticResources           *synthetic.Resources
	namespaces                   []string
	perturbationWaitDuration     time.Duration
	deletePoliciesBeforeTestCase bool
	verifyStateBeforeTestCase    bool
}

func NewInterpreter(kubernetes *kube.Kubernetes, namespaces []string, pods []string, ports []int, protocols []v1.Protocol, deletePoliciesBeforeTestCase bool, verifyStateBeforeTestCase bool) (*Interpreter, error) {
	kubeResources := connectivitykube.NewDefaultResources(namespaces, pods, ports, protocols)
	err := SetupCluster(kubernetes, kubeResources)
	if err != nil {
		return nil, err
	}
	syntheticResources, err := GetSyntheticResources(kubernetes, kubeResources)
	if err != nil {
		return nil, err
	}

	return &Interpreter{
		kubernetes:                   kubernetes,
		kubeResources:                kubeResources,
		syntheticResources:           syntheticResources,
		namespaces:                   namespaces,
		perturbationWaitDuration:     5 * time.Second, // TODO parameterize
		deletePoliciesBeforeTestCase: deletePoliciesBeforeTestCase,
		verifyStateBeforeTestCase:    verifyStateBeforeTestCase,
	}, nil
}

type Result struct {
	TestCase *generator.TestCase
	Steps    []*StepResult
	Err      error
}

type StepResult struct {
	SyntheticResult *synthetic.Result
	KubeResult      *connectivitykube.Results
	Policy          *matcher.Policy
	KubePolicies    []*networkingv1.NetworkPolicy
}

func (t *Interpreter) ExecuteTestCase(testCase *generator.TestCase) *Result {
	result := &Result{TestCase: testCase}
	var err error

	if t.deletePoliciesBeforeTestCase {
		// clean out all network policies
		err = t.kubernetes.DeleteAllNetworkPoliciesInNamespaces(t.namespaces)
		if err != nil {
			result.Err = err
			return result
		}
	}

	if t.verifyStateBeforeTestCase {
		err = t.verifyClusterState()
		if err != nil {
			result.Err = err
			return result
		}
		logrus.Info("cluster state verified")
	} else {
		logrus.Warnf("cluster state not verified")
	}

	// keep namespacesAndPods and kubePolicies in sync with what's in the cluster,
	//   so that we can correctly simulate expected results
	namespacesAndPods := t.syntheticResources
	var kubePolicies []*networkingv1.NetworkPolicy

	// perform perturbations one at a time, and run a probe after each change
	for _, step := range testCase.Steps {
		// TODO grab actual netpols from kube and record in results, for extra debugging/sanity checks

		for _, action := range step.Actions {
			if action.CreatePolicy != nil {
				// TODO blow up if it already exists?
				kubePolicy := action.CreatePolicy.Policy
				kubePolicies = append(kubePolicies, kubePolicy)
				_, err = t.kubernetes.CreateNetworkPolicy(kubePolicy)
				if err != nil {
					result.Err = err
					return result
				}
			} else if action.UpdatePodLabel != nil {
				update := action.UpdatePodLabel
				namespacesAndPods, err = namespacesAndPods.UpdatePodLabel(update.Namespace, update.Pod, update.Value, update.Key)
				if err != nil {
					result.Err = err
					return result
				}
				_, err = t.kubernetes.UpdatePodLabel(update.Namespace, update.Pod, update.Key, update.Value)
				if err != nil {
					result.Err = err
					return result
				}
			} else if action.ReadNetworkPolicies != nil {
				policies, err := t.kubernetes.GetNetworkPoliciesInNamespaces(action.ReadNetworkPolicies.Namespaces)
				if err != nil {
					result.Err = err
					return result
				}
				kubePolicies = append(kubePolicies, getSliceOfPointers(policies)...)
			} else {
				panic(errors.Errorf("invalid Action"))
			}
		}

		logrus.Infof("waiting %f seconds for perturbation to take affect", t.perturbationWaitDuration.Seconds())
		time.Sleep(t.perturbationWaitDuration)

		parsedPolicy := matcher.BuildNetworkPolicies(kubePolicies)

		logrus.Infof("running probe on port %d, protocol %s", step.Port, step.Protocol)

		stepResult := &StepResult{
			SyntheticResult: synthetic.RunSyntheticProbe(&synthetic.Request{
				Protocol:  step.Protocol,
				Port:      step.Port,
				Policies:  parsedPolicy,
				Resources: namespacesAndPods,
			}),
			KubeResult: connectivitykube.RunKubeProbe(t.kubernetes, &connectivitykube.Request{
				Resources:       t.kubeResources,
				Port:            step.Port,
				Protocol:        step.Protocol,
				NumberOfWorkers: 5,
			}),
			Policy:       parsedPolicy,
			KubePolicies: append([]*networkingv1.NetworkPolicy{}, kubePolicies...), // this looks weird, but just making a new copy to avoid accidentally mutating it elsewhere
		}
		result.Steps = append(result.Steps, stepResult)
	}

	return result
}

func getSliceOfPointers(netpols []networkingv1.NetworkPolicy) []*networkingv1.NetworkPolicy {
	netpolPointers := make([]*networkingv1.NetworkPolicy, len(netpols))
	for i := range netpols {
		netpolPointers[i] = &netpols[i]
	}
	return netpolPointers
}

func (t *Interpreter) verifyClusterState() error {
	kubePods, err := t.kubernetes.GetPodsInNamespaces(t.kubeResources.NamespacesSlice())
	if err != nil {
		return err
	}

	// 1. pods: labels, ips, containers, ports
	actualPods := map[string]v1.Pod{}
	for _, kubePod := range kubePods {
		actualPods[utils.NewPodString(kubePod.Namespace, kubePod.Name).String()] = kubePod
	}
	// are we missing any pods?
	for _, pod := range t.syntheticResources.Pods {
		if actualPod, ok := actualPods[pod.PodString().String()]; ok {
			if !areLabelsEqual(actualPod.Labels, pod.Labels) {
				return errors.Errorf("for pod %s, expected labels %+v (found %+v)", pod.PodString().String(), pod.Labels, actualPod.Labels)
			}
			if actualPod.Status.PodIP != pod.IP {
				return errors.Errorf("for pod %s, expected ip %s (found %s)", pod.PodString().String(), pod.IP, actualPod.Status.PodIP)
			}
			if !areContainersEqual(actualPod, pod) {
				return errors.Errorf("for pod %s, expected containers %+v (found %+v)", pod.PodString().String(), pod.Containers, actualPod.Spec.Containers)
			}
		} else {
			return errors.Errorf("missing expected pod %s", pod.PodString().String())
		}
	}

	// 2. services: selectors, ports
	for _, pod := range t.kubeResources.Pods {
		expected := pod.KubeService
		svc, err := t.kubernetes.GetService(expected.Namespace, expected.Name)
		if err != nil {
			return err
		}
		if !areLabelsEqual(svc.Spec.Selector, pod.Labels) {
			return errors.Errorf("for service %s/%s, expected labels %+v (found %+v)", pod.Namespace, pod.Name, pod.Labels, svc.Spec.Selector)
		}
		if len(expected.Spec.Ports) != len(svc.Spec.Ports) {
			return errors.Errorf("for service %s/%s, expected %d ports (found %d)", expected.Namespace, expected.Name, len(expected.Spec.Ports), len(svc.Spec.Ports))
		}
		for i, port := range expected.Spec.Ports {
			kubePort := svc.Spec.Ports[i]
			if kubePort.Protocol != port.Protocol || kubePort.Port != port.Port {
				return errors.Errorf("for service %s/%s, expected port %+v (found %+v)", expected.Namespace, expected.Name, port, kubePort)
			}
		}
	}

	// 3. namespaces: names, labels
	for ns, labels := range t.syntheticResources.Namespaces {
		namespace, err := t.kubernetes.GetNamespace(ns)
		if err != nil {
			return err
		}
		if !areLabelsEqual(namespace.Labels, labels) {
			return errors.Errorf("for namespace %s, expected labels %+v (found %+v)", ns, labels, namespace.Labels)
		}
	}

	// nothing wrong: we're good to go
	return nil
}

func areContainersEqual(kubePod v1.Pod, expectedPod *synthetic.Pod) bool {
	kubeConts := kubePod.Spec.Containers
	if len(kubeConts) != len(expectedPod.Containers) {
		return false
	}
	for i, kubeCont := range kubeConts {
		cont := expectedPod.Containers[i]
		if len(kubeCont.Ports) != 1 {
			return false
		}
		if int(kubeCont.Ports[0].ContainerPort) != cont.Port {
			return false
		}
		if kubeCont.Ports[0].Protocol != cont.Protocol {
			return false
		}
	}

	return true
}

func areLabelsEqual(l map[string]string, r map[string]string) bool {
	if len(l) != len(r) {
		return false
	}
	for k, lv := range l {
		rv, ok := r[k]
		if !ok || lv != rv {
			return false
		}
	}
	return true
}
