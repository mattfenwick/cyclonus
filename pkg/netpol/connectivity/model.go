package connectivity

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Model defines the namespaces, deployments, services, pods, containers and associated
// data for network policy test cases and provides the source of truth
type Model struct {
	Namespaces    []*Namespace
	allPodStrings *[]PodString
	allPods       *[]*Pod
	// the raw data
	NamespaceNames []string
	PodNames       []string
	Ports          []int32
	Protocols      []v1.Protocol
	DNSDomain      string
}

// NewModel instantiates a model based on:
// - namespaces
// - pods
// The total number of pods is the number of namespaces x the number of pods per namespace.
func NewModel(namespaces []string, podNames []string) *Model {
	model := &Model{
		NamespaceNames: namespaces,
		PodNames:       podNames,
	}

	// build the entire "model" for the overall test, which means, building
	// namespaces, pods, containers for each protocol.
	for _, ns := range namespaces {
		var pods []*Pod
		for _, podName := range podNames {
			pods = append(pods, &Pod{
				Namespace: ns,
				Name:      podName,
				// TODO do something different with these labels -- make them injectable?
				Labels: map[string]string{"pod": podName},
			})
		}
		model.Namespaces = append(
			model.Namespaces,
			// TODO do something different with these labels -- make them injectable?
			&Namespace{Name: ns, Pods: pods, Labels: map[string]string{"ns": ns}})
	}
	return model
}

// NewTruthTable instantiates a default-true truth table
func (m *Model) NewTruthTable() *TruthTable {
	var podNames []string
	for _, pod := range m.AllPods() {
		podNames = append(podNames, pod.PodString().String())
	}
	return NewTruthTableFromItems(podNames, nil)
}

// AllPodStrings returns a slice of all pod strings
func (m *Model) AllPodStrings() []PodString {
	if m.allPodStrings == nil {
		var pods []PodString
		for _, ns := range m.Namespaces {
			for _, pod := range ns.Pods {
				pods = append(pods, pod.PodString())
			}
		}
		m.allPodStrings = &pods
	}
	return *m.allPodStrings
}

// AllPods returns a slice of all pods
func (m *Model) AllPods() []*Pod {
	if m.allPods == nil {
		var pods []*Pod
		for _, ns := range m.Namespaces {
			for _, pod := range ns.Pods {
				pods = append(pods, pod)
			}
		}
		m.allPods = &pods
	}
	return *m.allPods
}

// Namespace is the abstract representation of what matters to network policy
// tests for a namespace; i.e. it ignores kube implementation details
type Namespace struct {
	Name   string
	Pods   []*Pod
	Labels map[string]string
}

// Spec builds a kubernetes namespace spec
func (ns *Namespace) Spec() *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ns.Name,
			Labels: ns.LabelSelector(),
		},
	}
}

// LabelSelector returns the default labels that should be placed on a namespace
// in order for it to be uniquely selectable by label selectors
func (ns *Namespace) LabelSelector() map[string]string {
	return map[string]string{"ns": ns.Name}
}

// Pod is the abstract representation of what matters to network policy tests for
// a pod; i.e. it ignores kube implementation details
type Pod struct {
	Namespace string
	Name      string
	Labels    map[string]string
}

// PodString returns a corresponding pod string
func (p *Pod) PodString() PodString {
	return NewPodString(p.Namespace, p.Name)
}
