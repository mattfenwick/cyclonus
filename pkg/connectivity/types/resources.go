package types

import (
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
)

const (
	agnhostImage = "k8s.gcr.io/e2e-test-images/agnhost:2.21"
)

type Resources struct {
	Namespaces  map[string]map[string]string
	Pods        []*Pod
	ExternalIPs []string
}

func NewDefaultResources(namespaces []string, podNames []string, ports []int, protocols []v1.Protocol, externalIPs []string) *Resources {
	sort.Strings(externalIPs)
	r := &Resources{
		Namespaces:  map[string]map[string]string{},
		ExternalIPs: externalIPs,
	}

	for _, ns := range namespaces {
		for _, podName := range podNames {
			r.Pods = append(r.Pods, NewDefaultPod(ns, podName, map[string]string{"pod": podName}, "TODO", ports, protocols))
		}
		r.Namespaces[ns] = map[string]string{"ns": ns}
	}

	return r
}

func NewResources(namespaces map[string]map[string]string, pods []*Pod, externalIPs []string) (*Resources, error) {
	sort.Strings(externalIPs)
	model := &Resources{Namespaces: namespaces, ExternalIPs: externalIPs}

	for _, pod := range pods {
		if _, ok := namespaces[pod.Namespace]; !ok {
			return nil, errors.Errorf("namespace for pod %s/%s not found", pod.Namespace, pod.Name)
		}
		model.Pods = append(model.Pods, pod)
	}

	return model, nil
}

func NewResourcesFromKube(kubernetes *kube.Kubernetes, namespaces []string) (*Resources, error) {
	podList, err := kubernetes.GetPodsInNamespaces(namespaces)
	if err != nil {
		return nil, err
	}
	var pods []*Pod
	for _, pod := range podList {
		ip := pod.Status.PodIP
		if ip == "" {
			return nil, errors.Errorf("no ip found for pod %s/%s", pod.Namespace, pod.Name)
		}
		var containers []*Container
		for _, kubeCont := range pod.Spec.Containers {
			if len(kubeCont.Ports) != 1 {
				return nil, errors.Errorf("expected 1 port on kube container, found %d", len(kubeCont.Ports))
			}
			kubePort := kubeCont.Ports[0]
			containers = append(containers, &Container{
				Name:     kubeCont.Name,
				Port:     int(kubePort.ContainerPort),
				Protocol: kubePort.Protocol,
				PortName: kubePort.Name,
			})
		}
		pods = append(pods, &Pod{
			Namespace:  pod.Namespace,
			Name:       pod.Name,
			Labels:     pod.Labels,
			IP:         ip,
			Containers: containers,
		})
		logrus.Debugf("ip for pod %s/%s: %s", pod.Namespace, pod.Name, ip)
	}

	kubeNamespaces := map[string]map[string]string{}
	for _, ns := range namespaces {
		kubeNs, err := kubernetes.GetNamespace(ns)
		if err != nil {
			return nil, err
		}
		kubeNamespaces[ns] = kubeNs.Labels
	}

	return &Resources{
		Namespaces:  kubeNamespaces,
		Pods:        pods,
		ExternalIPs: nil,
	}, nil
}

// UpdateNamespaceLabels returns a new object with an updated namespace.  It should not affect the original Resources object.
func (r *Resources) UpdateNamespaceLabels(ns string, labels map[string]string) (*Resources, error) {
	if _, ok := r.Namespaces[ns]; !ok {
		return nil, errors.Errorf("no namespace %s", ns)
	}
	newNamespaces := map[string]map[string]string{}
	for oldNs, oldLabels := range r.Namespaces {
		newNamespaces[oldNs] = oldLabels
	}
	newNamespaces[ns] = labels
	return &Resources{
		Namespaces: newNamespaces,
		Pods:       r.Pods,
	}, nil
}

// UpdatePodLabel returns a new object with an updated pod.  It should not affect the original Resources object.
func (r *Resources) SetPodLabels(ns string, podName string, labels map[string]string) (*Resources, error) {
	var pods []*Pod
	found := false
	for _, existingPod := range r.Pods {
		if existingPod.Namespace == ns && existingPod.Name == podName {
			found = true
			pods = append(pods, existingPod.SetLabels(labels))
		} else {
			pods = append(pods, existingPod)
		}
	}
	if !found {
		return nil, errors.Errorf("no pod named %s/%s found", ns, podName)
	}
	return &Resources{
		Namespaces:  r.Namespaces,
		Pods:        pods,
		ExternalIPs: r.ExternalIPs,
	}, nil
}

func (r *Resources) NewTable() *Table {
	var podNames []string
	for _, pod := range r.Pods {
		podNames = append(podNames, pod.PodString().String())
	}
	sort.Strings(podNames)
	return NewTable(append(podNames, r.ExternalIPs...))
}

func (r *Resources) NamespacesSlice() []string {
	var nss []string
	for ns := range r.Namespaces {
		nss = append(nss, ns)
	}
	return nss
}

func (r *Resources) CreateResourcesInKube(kube *kube.Kubernetes) error {
	for ns, labels := range r.Namespaces {
		_, err := kube.CreateOrUpdateNamespace(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns, Labels: labels}})
		if err != nil {
			return err
		}
	}
	for _, pod := range r.Pods {
		_, err := kube.CreatePodIfNotExists(pod.KubePod())
		if err != nil {
			return err
		}
		_, err = kube.CreateServiceIfNotExists(pod.KubeService())
		if err != nil {
			return err
		}
	}
	return nil
}
