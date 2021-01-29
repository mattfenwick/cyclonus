package synthetic

import (
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"sort"
)

type Resources struct {
	Namespaces map[string]map[string]string
	Pods       []*Pod
}

func NewResources(namespaces map[string]map[string]string, pods []*Pod) (*Resources, error) {
	model := &Resources{Namespaces: namespaces}

	for _, pod := range pods {
		if _, ok := namespaces[pod.Namespace]; !ok {
			return nil, errors.Errorf("namespace for pod %s/%s not found", pod.Namespace, pod.Name)
		}
		model.Pods = append(model.Pods, pod)
	}

	return model, nil
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
		Namespaces: r.Namespaces,
		Pods:       pods,
	}, nil
}

func (r *Resources) NewTruthTable() *utils.TruthTable {
	var podNames []string
	for _, pod := range r.Pods {
		podNames = append(podNames, pod.PodString().String())
	}
	sort.Slice(podNames, func(i, j int) bool {
		return podNames[i] < podNames[j]
	})
	return utils.NewTruthTableFromItems(podNames, nil)
}

type Namespace struct {
	Labels map[string]string
}

type Pod struct {
	Namespace  string
	Name       string
	Labels     map[string]string
	IP         string
	Containers []*Container
}

func (p *Pod) SetLabels(labels map[string]string) *Pod {
	return &Pod{
		Namespace: p.Namespace,
		Name:      p.Name,
		Labels:    labels,
		IP:        p.IP,
	}
}

func (p *Pod) PodString() utils.PodString {
	return utils.NewPodString(p.Namespace, p.Name)
}

type Container struct {
	Port     int
	Protocol v1.Protocol
}
