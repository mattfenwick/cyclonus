package connectivity

import (
	connectivitykube "github.com/mattfenwick/cyclonus/pkg/connectivity/kube"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/synthetic"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
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
	port                         int
	protocol                     v1.Protocol
	deletePoliciesBeforeTestCase bool
	verifyStateBeforeTestCase    bool
	performPreProbe              bool
}

func NewInterpreter(kubernetes *kube.Kubernetes, namespaces []string, pods []string, port int, protocol v1.Protocol, deletePoliciesBeforeTestCase bool, verifyStateBeforeTestCase bool, performPreProbe bool) (*Interpreter, error) {
	kubeResources := connectivitykube.NewDefaultResources(namespaces, pods, port, protocol)
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
		port:                         port,
		protocol:                     protocol,
		deletePoliciesBeforeTestCase: deletePoliciesBeforeTestCase,
		verifyStateBeforeTestCase:    verifyStateBeforeTestCase,
		performPreProbe:              performPreProbe,
	}, nil
}

type Result struct {
	TestCase *generator.TestCase
	PreProbe *StepResult
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
		// TODO make sure cluster matches t.kubeResources:
		//   update namespaces and pods to match
		//   blow up if any pod IPs are different (since the network policies may depend on specific ips)
	}

	// keep namespacesAndPods and kubePolicies in sync with what's in the cluster,
	//   so that we can correctly simulate expected results
	namespacesAndPods := t.syntheticResources
	var kubePolicies []*networkingv1.NetworkPolicy

	if t.performPreProbe {
		// run a pre-probe to make sure everything is in order before performing any perturbations
		result.PreProbe = &StepResult{
			SyntheticResult: synthetic.RunSyntheticProbe(&synthetic.Request{
				Protocol:  t.protocol,
				Port:      t.port,
				Policies:  matcher.BuildNetworkPolicies(kubePolicies),
				Resources: namespacesAndPods,
			}),
			KubeResult: connectivitykube.RunKubeProbe(t.kubernetes, &connectivitykube.Request{
				Resources:       t.kubeResources,
				Port:            t.port,
				Protocol:        t.protocol,
				NumberOfWorkers: 5,
			}),
			Policy: matcher.BuildNetworkPolicies(kubePolicies),
		}
	}

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

		logrus.Infof("running probe on port %d, protocol %s", t.port, t.protocol)

		stepResult := &StepResult{
			SyntheticResult: synthetic.RunSyntheticProbe(&synthetic.Request{
				Protocol:  t.protocol,
				Port:      t.port,
				Policies:  parsedPolicy,
				Resources: namespacesAndPods,
			}),
			KubeResult: connectivitykube.RunKubeProbe(t.kubernetes, &connectivitykube.Request{
				Resources:       t.kubeResources,
				Port:            t.port,
				Protocol:        t.protocol,
				NumberOfWorkers: 5,
			}),
			Policy:       parsedPolicy,
			KubePolicies: append([]*networkingv1.NetworkPolicy{}, kubePolicies...),
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
