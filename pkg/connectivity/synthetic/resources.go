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

func (m *Resources) NewTruthTable() *utils.TruthTable {
	var podNames []string
	for _, pod := range m.Pods {
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

func (p *Pod) PodString() utils.PodString {
	return utils.NewPodString(p.Namespace, p.Name)
}
