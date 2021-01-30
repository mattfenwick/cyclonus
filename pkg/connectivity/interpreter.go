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
	kubernetes                 *kube.Kubernetes
	kubeResources              *connectivitykube.Resources
	syntheticResources         *synthetic.Resources
	namespaces                 []string
	kubeProbeRetries           int
	perturbationWaitDuration   time.Duration
	resetClusterBeforeTestCase bool
}

func NewInterpreter(kubernetes *kube.Kubernetes, namespaces []string, pods []string, ports []int, protocols []v1.Protocol, resetClusterBeforeTestCase bool, kubeProbeRetries int) (*Interpreter, error) {
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
		kubernetes:                 kubernetes,
		kubeResources:              kubeResources,
		syntheticResources:         syntheticResources,
		namespaces:                 namespaces,
		perturbationWaitDuration:   5 * time.Second, // TODO parameterize
		resetClusterBeforeTestCase: resetClusterBeforeTestCase,
		kubeProbeRetries:           kubeProbeRetries,
	}, nil
}

type Result struct {
	TestCase *generator.TestCase
	Steps    []*StepResult
	Err      error
}

type StepResult struct {
	SyntheticResult *synthetic.Result
	KubeResults     []*connectivitykube.Results
	Policy          *matcher.Policy
	KubePolicies    []*networkingv1.NetworkPolicy
}

func (s *StepResult) LastKubeResult() *connectivitykube.Results {
	return s.KubeResults[len(s.KubeResults)-1]
}

func (t *Interpreter) ExecuteTestCase(testCase *generator.TestCase) *Result {
	result := &Result{TestCase: testCase}

	if t.resetClusterBeforeTestCase {
		err := t.resetClusterState()
		if err != nil {
			result.Err = err
			return result
		}
	}

	err := t.verifyClusterState()
	if err != nil {
		result.Err = err
		return result
	}
	logrus.Info("cluster state verified")

	// keep namespacesAndPods and kubePolicies in sync with what's in the cluster,
	//   so that we can correctly simulate expected results
	namespacesAndPods := t.syntheticResources
	var kubePolicies []*networkingv1.NetworkPolicy

	// perform perturbations one at a time, and run a probe after each change
	for stepIndex, step := range testCase.Steps {
		// TODO grab actual netpols from kube and record in results, for extra debugging/sanity checks

		for _, action := range step.Actions {
			if action.CreatePolicy != nil {
				newPolicy := action.CreatePolicy.Policy
				// do we already have this policy?
				for _, kubePol := range kubePolicies {
					if kubePol.Namespace == newPolicy.Namespace && kubePol.Name == newPolicy.Name {
						result.Err = errors.Errorf("cannot create policy %s/%s: already exists", newPolicy.Namespace, newPolicy.Name)
						return result
					}
				}

				kubePolicies = append(kubePolicies, newPolicy)
				_, err = t.kubernetes.CreateNetworkPolicy(newPolicy)
				if err != nil {
					result.Err = err
					return result
				}
			} else if action.UpdatePolicy != nil {
				newPolicy := action.UpdatePolicy.Policy
				// we already have this policy -- right?
				index := -1
				found := false
				for i, kubePol := range kubePolicies {
					if kubePol.Namespace == newPolicy.Namespace && kubePol.Name == newPolicy.Name {
						index = i
						found = true
						break
					}
				}
				if !found {
					result.Err = errors.Errorf("cannot update policy %s/%s: not found", newPolicy.Namespace, newPolicy.Name)
					return result
				}

				kubePolicies[index] = newPolicy
				_, err = t.kubernetes.UpdateNetworkPolicy(newPolicy)
				if err != nil {
					result.Err = err
					return result
				}
			} else if action.SetNamespaceLabels != nil {
				newNs := action.SetNamespaceLabels.Namespace
				newLabels := action.SetNamespaceLabels.Labels
				namespacesAndPods, err = namespacesAndPods.UpdateNamespaceLabels(newNs, newLabels)
				if err != nil {
					result.Err = err
					return result
				}
				_, err = t.kubernetes.SetNamespaceLabels(newNs, newLabels)
				if err != nil {
					result.Err = err
					return result
				}
			} else if action.SetPodLabels != nil {
				update := action.SetPodLabels
				namespacesAndPods, err = namespacesAndPods.SetPodLabels(update.Namespace, update.Pod, update.Labels)
				if err != nil {
					result.Err = err
					return result
				}
				_, err = t.kubernetes.SetPodLabels(update.Namespace, update.Pod, update.Labels)
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
			} else if action.DeletePolicy != nil {
				ns := action.DeletePolicy.Namespace
				name := action.DeletePolicy.Name
				// make sure this policy exists
				index := -1
				found := false
				for i, kubePol := range kubePolicies {
					if kubePol.Namespace == ns && kubePol.Name == name {
						found = true
						index = i
					}
				}
				if !found {
					result.Err = errors.Errorf("cannot delete policy %s/%s: not found", ns, name)
					return result
				}

				var newPolicies []*networkingv1.NetworkPolicy
				for i, kubePol := range kubePolicies {
					if i != index {
						newPolicies = append(newPolicies, kubePol)
					}
				}
				kubePolicies = newPolicies

				err = t.kubernetes.DeleteNetworkPolicy(ns, name)
				if err != nil {
					result.Err = err
					return result
				}
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
			Policy:       parsedPolicy,
			KubePolicies: append([]*networkingv1.NetworkPolicy{}, kubePolicies...), // this looks weird, but just making a new copy to avoid accidentally mutating it elsewhere
		}

		for i := 0; i <= t.kubeProbeRetries; i++ {
			logrus.Infof("running kube probe on step %d, try %d", stepIndex, i)
			kubeProbe := connectivitykube.RunKubeProbe(t.kubernetes, &connectivitykube.Request{
				Resources:       t.kubeResources,
				Port:            step.Port,
				Protocol:        step.Protocol,
				NumberOfWorkers: 5,
			})
			stepResult.KubeResults = append(stepResult.KubeResults, kubeProbe)
			if counts := kubeProbe.TruthTable().Compare(stepResult.SyntheticResult.Combined).ValueCounts(false); counts.False == 0 {
				break
			}
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

func (t *Interpreter) resetClusterState() error {
	err := t.kubernetes.DeleteAllNetworkPoliciesInNamespaces(t.namespaces)
	if err != nil {
		return err
	}

	for ns, labels := range t.syntheticResources.Namespaces {
		_, err = t.kubernetes.SetNamespaceLabels(ns, labels)
		if err != nil {
			return err
		}
	}

	for _, pod := range t.syntheticResources.Pods {
		_, err = t.kubernetes.SetPodLabels(pod.Namespace, pod.Name, pod.Labels)
		if err != nil {
			return err
		}
	}

	//for _, step := range steps {
	//	for _, action := range step.Actions {
	//		if action.CreatePolicy != nil {
	//			// nothing to do
	//		} else if action.UpdatePolicy != nil {
	//			// nothing to do
	//		} else if action.SetNamespaceLabels != nil {
	//			newNs := action.SetNamespaceLabels.Namespace
	//			expectedLabels := t.syntheticResources.Namespaces[newNs]
	//			_, err := t.kubernetes.SetNamespaceLabels(newNs, expectedLabels)
	//			if err != nil {
	//				return err
	//			}
	//		} else if action.SetPodLabels != nil {
	//			update := action.SetPodLabels
	//			var pod *synthetic.Pod
	//			for _, p := range t.syntheticResources.Pods {
	//				if p.Namespace == update.Namespace && p.Name == update.Pod {
	//					pod = p
	//				}
	//			}
	//			if pod == nil {
	//				return errors.Errorf("pod %s/%s not found", update.Namespace, update.Pod)
	//			}
	//			_, err := t.kubernetes.SetPodLabels(update.Namespace, update.Pod, pod.Labels)
	//			if err != nil {
	//				return err
	//			}
	//		} else if action.ReadNetworkPolicies != nil {
	//			// nothing to do
	//		} else if action.DeletePolicy != nil {
	//			// nothing to do
	//		} else {
	//			panic(errors.Errorf("invalid Action"))
	//		}
	//	}
	//}
	return nil
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

	// 4. network policies
	policies, err := t.kubernetes.GetNetworkPoliciesInNamespaces(t.namespaces)
	if err != nil {
		return err
	}
	if len(policies) > 0 {
		return errors.Errorf("expected 0 policies in namespaces %+v, found %d", t.namespaces, len(policies))
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
