package connectivity

import (
	"github.com/mattfenwick/cyclonus/pkg/connectivity/probe"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"time"
)

type TestCaseState struct {
	Kubernetes kube.IKubernetes
	Resources  *probe.Resources
	Policies   []*networkingv1.NetworkPolicy
}

func (t *TestCaseState) CreatePolicy(policy *networkingv1.NetworkPolicy) error {
	// do we already have this policy?
	for _, kubePol := range t.Policies {
		if kubePol.Namespace == policy.Namespace && kubePol.Name == policy.Name {
			return errors.Errorf("cannot create policy %s/%s: already exists", policy.Namespace, policy.Name)
		}
	}
	t.Policies = append(t.Policies, policy)

	_, err := t.Kubernetes.CreateNetworkPolicy(policy)
	return err
}

func (t *TestCaseState) UpdatePolicy(policy *networkingv1.NetworkPolicy) error {
	// we already have this policy -- right?
	index := -1
	found := false
	for i, kubePol := range t.Policies {
		if kubePol.Namespace == policy.Namespace && kubePol.Name == policy.Name {
			index = i
			found = true
			break
		}
	}
	if !found {
		return errors.Errorf("cannot update policy %s/%s: not found", policy.Namespace, policy.Name)
	}

	t.Policies[index] = policy
	_, err := t.Kubernetes.UpdateNetworkPolicy(policy)
	return err
}

func (t *TestCaseState) CreateNamespace(ns string, labels map[string]string) error {
	newResources, err := t.Resources.CreateNamespace(ns, labels)
	if err != nil {
		return err
	}
	t.Resources = newResources
	_, err = t.Kubernetes.CreateNamespace(probe.KubeNamespace(ns, labels))
	return err
}

func (t *TestCaseState) SetNamespaceLabels(ns string, labels map[string]string) error {
	newResources, err := t.Resources.UpdateNamespaceLabels(ns, labels)
	if err != nil {
		return err
	}
	t.Resources = newResources
	_, err = t.Kubernetes.SetNamespaceLabels(ns, labels)
	return err
}

func (t *TestCaseState) DeleteNamespace(ns string) error {
	newResources, err := t.Resources.DeleteNamespace(ns)
	if err != nil {
		return err
	}
	t.Resources = newResources
	return t.Kubernetes.DeleteNamespace(ns)
}

func (t *TestCaseState) CreatePod(ns string, pod string, labels map[string]string) error {
	newResources, err := t.Resources.CreatePod(ns, pod, labels)
	if err != nil {
		return err
	}
	t.Resources = newResources
	newPod, err := newResources.GetPod(ns, pod)
	if err != nil {
		return err
	}
	_, err = t.Kubernetes.CreatePod(newPod.KubePod())
	if err != nil {
		return err
	}
	_, err = t.Kubernetes.CreateService(newPod.KubeService())
	if err != nil {
		return err
	}
	// wait for ready, get ip
	for i := 0; i < 12; i++ {
		kubePod, err := t.Kubernetes.GetPod(ns, pod)
		if err != nil {
			return err
		}
		if kubePod.Status.Phase == "Running" && kubePod.Status.PodIP != "" {
			newPod.IP = kubePod.Status.PodIP
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return errors.Errorf("unable to wait for running or get pod ip for %s/%s after creation", ns, pod)
}

func (t *TestCaseState) SetPodLabels(ns string, pod string, labels map[string]string) error {
	newResources, err := t.Resources.SetPodLabels(ns, pod, labels)
	if err != nil {
		return err
	}
	t.Resources = newResources
	_, err = t.Kubernetes.SetPodLabels(ns, pod, labels)
	return err
}

func (t *TestCaseState) DeletePod(ns string, pod string) error {
	deletedPod, err := t.Resources.GetPod(ns, pod)
	if err != nil {
		return err
	}
	newResources, err := t.Resources.DeletePod(ns, pod)
	if err != nil {
		return err
	}
	t.Resources = newResources
	err = t.Kubernetes.DeleteService(ns, deletedPod.KubeService().Name)
	if err != nil {
		return err
	}
	return t.Kubernetes.DeletePod(ns, pod)
}

func (t *TestCaseState) ReadPolicies(namespaces []string) error {
	policies, err := kube.GetNetworkPoliciesInNamespaces(t.Kubernetes, namespaces)
	if err != nil {
		return err
	}
	t.Policies = append(t.Policies, getSliceOfPointers(policies)...)
	return nil
}

func (t *TestCaseState) DeletePolicy(ns string, name string) error {
	// make sure this policy exists
	index := -1
	found := false
	for i, kubePol := range t.Policies {
		if kubePol.Namespace == ns && kubePol.Name == name {
			found = true
			index = i
		}
	}
	if !found {
		return errors.Errorf("cannot delete policy %s/%s: not found", ns, name)
	}

	var newPolicies []*networkingv1.NetworkPolicy
	for i, kubePol := range t.Policies {
		if i != index {
			newPolicies = append(newPolicies, kubePol)
		}
	}
	t.Policies = newPolicies

	return t.Kubernetes.DeleteNetworkPolicy(ns, name)
}

func getSliceOfPointers(netpols []networkingv1.NetworkPolicy) []*networkingv1.NetworkPolicy {
	netpolPointers := make([]*networkingv1.NetworkPolicy, len(netpols))
	for i := range netpols {
		netpolPointers[i] = &netpols[i]
	}
	return netpolPointers
}

func (t *TestCaseState) verifyClusterStateHelper() error {
	kubePods, err := kube.GetPodsInNamespaces(t.Kubernetes, t.Resources.NamespacesSlice())
	if err != nil {
		return err
	}

	// 1. pods: labels, ips, containers, ports
	actualPods := map[string]v1.Pod{}
	for _, kubePod := range kubePods {
		actualPods[probe.NewPodString(kubePod.Namespace, kubePod.Name).String()] = kubePod
	}
	// are we missing any pods?
	for _, expectedPod := range t.Resources.Pods {
		if actualPod, ok := actualPods[expectedPod.PodString().String()]; ok {
			if !NewLabelsDiff(actualPod.Labels, expectedPod.Labels).AreLabelsEqual() {
				return errors.Errorf("for pod %s, expected labels %+v (found %+v)", expectedPod.PodString().String(), expectedPod.Labels, actualPod.Labels)
			}
			if actualPod.Status.PodIP != expectedPod.IP {
				return errors.Errorf("for pod %s, expected ip %s (found %s)", expectedPod.PodString().String(), expectedPod.IP, actualPod.Status.PodIP)
			}
			if !expectedPod.IsEqualToKubePod(actualPod) {
				return errors.Errorf("for pod %s, expected containers %+v (found %+v)", expectedPod.PodString().String(), expectedPod.Containers, actualPod.Spec.Containers)
			}
		} else {
			return errors.Errorf("missing expected pod %s", expectedPod.PodString().String())
		}
	}

	// 2. services: selectors, ports
	for _, expectedPod := range t.Resources.Pods {
		expected := expectedPod.KubeService()
		svc, err := t.Kubernetes.GetService(expected.Namespace, expected.Name)
		if err != nil {
			return err
		}
		if !NewLabelsDiff(svc.Spec.Selector, expectedPod.Labels).AreLabelsEqual() {
			return errors.Errorf("for service %s/%s, expected labels %+v (found %+v)", expectedPod.Namespace, expectedPod.Name, expectedPod.Labels, svc.Spec.Selector)
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
	for ns, expectedNamespaceLabels := range t.Resources.Namespaces {
		namespace, err := t.Kubernetes.GetNamespace(ns)
		if err != nil {
			return err
		}
		if !NewLabelsDiff(namespace.Labels, expectedNamespaceLabels).AreLabelsEqual() {
			return errors.Errorf("for namespace %s, expected labels %+v (found %+v)", ns, expectedNamespaceLabels, namespace.Labels)
		}
	}

	// nothing wrong: we're good to go
	return nil
}

type LabelsDiff struct {
	Same      []string
	Different []string
	Extra     []string
	Missing   []string
}

func NewLabelsDiff(actual map[string]string, expected map[string]string) *LabelsDiff {
	ld := &LabelsDiff{
		Same:      nil,
		Different: nil,
		Extra:     nil,
		Missing:   nil,
	}
	for k, actualValue := range actual {
		expectedValue, ok := expected[k]
		if !ok {
			ld.Extra = append(ld.Extra, k)
		} else if actualValue != expectedValue {
			ld.Different = append(ld.Different, k)
		} else {
			ld.Same = append(ld.Same, k)
		}
	}
	for k, _ := range expected {
		if _, ok := actual[k]; !ok {
			ld.Missing = append(ld.Missing, k)
		}
	}
	return ld
}

func (ld *LabelsDiff) AreLabelsEqual() bool {
	return len(ld.Different) == 0 && len(ld.Extra) == 0 && len(ld.Missing) == 0
}

func (t *TestCaseState) resetLabelsInKubeHelper() error {
	for ns, labels := range t.Resources.Namespaces {
		_, err := t.Kubernetes.SetNamespaceLabels(ns, labels)
		if err != nil {
			return err
		}
	}

	for _, pod := range t.Resources.Pods {
		_, err := t.Kubernetes.SetPodLabels(pod.Namespace, pod.Name, pod.Labels)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TestCaseState) ResetClusterState() error {
	err := kube.DeleteAllNetworkPoliciesInNamespaces(t.Kubernetes, t.Resources.NamespacesSlice())
	if err != nil {
		return err
	}

	return t.resetLabelsInKubeHelper()
}

func (t *TestCaseState) VerifyClusterState() error {
	err := t.verifyClusterStateHelper()
	if err != nil {
		return err
	}

	policies, err := kube.GetNetworkPoliciesInNamespaces(t.Kubernetes, t.Resources.NamespacesSlice())
	if err != nil {
		return err
	}
	if len(policies) > 0 {
		return errors.Errorf("expected 0 policies in namespaces %+v, found %d", t.Resources.NamespacesSlice(), len(policies))
	}
	return nil
}
