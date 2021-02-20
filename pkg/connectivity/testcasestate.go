package connectivity

import (
	"github.com/mattfenwick/cyclonus/pkg/connectivity/probe"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	networkingv1 "k8s.io/api/networking/v1"
	"time"
)

type TestCaseState struct {
	Kubernetes *kube.Kubernetes
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
	_, err = t.Kubernetes.CreateOrUpdateNamespace(probe.KubeNamespace(ns, labels))
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
		time.Sleep(5 * time.Second)
		kubePod, err := t.Kubernetes.GetPod(ns, pod)
		if err != nil {
			return err
		}
		if kubePod.Status.Phase == "Running" && kubePod.Status.PodIP != "" {
			newPod.IP = kubePod.Status.PodIP
			return nil
		}
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
	policies, err := t.Kubernetes.GetNetworkPoliciesInNamespaces(namespaces)
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
