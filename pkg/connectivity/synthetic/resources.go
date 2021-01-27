package synthetic

import (
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
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

// UpdatePodLabel returns a new object with an updated pod.  It should not affect the original Resources object.
func (r *Resources) UpdatePodLabel(ns string, podName string, key string, value string) (*Resources, error) {
	var pods []*Pod
	found := false
	for _, existingPod := range r.Pods {
		if existingPod.Namespace == ns && existingPod.Name == podName {
			found = true
			pods = append(pods, existingPod.UpdateLabel(key, value))
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
	Namespace string
	Name      string
	Labels    map[string]string
	IP        string
}

func (p *Pod) UpdateLabel(key string, value string) *Pod {
	labels := map[string]string{}
	for k, v := range p.Labels {
		labels[k] = v
	}
	labels[key] = value
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
