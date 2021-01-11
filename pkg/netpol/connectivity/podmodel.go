package connectivity

import "sort"

// PodModel defines the namespaces, deployments, services, pods, containers and associated
// data for network policy test cases and provides the source of truth
type PodModel struct {
	Namespaces map[string]*Namespace
	// derived
	allPodStrings *[]PodString
	allPods       *[]*NamespacedPod
}

// NewDefaultModel instantiates a model based on:
// - namespaces
// - pods
// The total number of pods is the number of namespaces x the number of pods per namespace.
func NewDefaultModel(namespaces []string, podNames []string) *PodModel {
	model := &PodModel{Namespaces: map[string]*Namespace{}}

	// build the entire "model" for the overall test, which means, building
	// namespaces, pods, containers for each protocol.
	for _, ns := range namespaces {
		pods := map[string]*Pod{}
		for _, podName := range podNames {
			pods[podName] = &Pod{
				Labels: map[string]string{"pod": podName},
			}
		}
		model.Namespaces[ns] = &Namespace{Pods: pods, Labels: map[string]string{"ns": ns}}
	}
	return model
}

// NewTruthTable instantiates a default-true truth table
func (m *PodModel) NewTruthTable() *TruthTable {
	var podNames []string
	for _, pod := range m.AllPodStrings() {
		podNames = append(podNames, pod.String())
	}
	sort.Slice(podNames, func(i, j int) bool {
		return podNames[i] < podNames[j]
	})
	return NewTruthTableFromItems(podNames, nil)
}

// AllPodStrings returns a slice of all pod strings
func (m *PodModel) AllPodStrings() []PodString {
	if m.allPodStrings == nil {
		var pods []PodString
		for nsName, ns := range m.Namespaces {
			for podName := range ns.Pods {
				pods = append(pods, NewPodString(nsName, podName))
			}
		}
		m.allPodStrings = &pods
	}
	return *m.allPodStrings
}

// AllPods returns a slice of all pods
func (m *PodModel) AllPods() []*NamespacedPod {
	if m.allPods == nil {
		var pods []*NamespacedPod
		for nsName, ns := range m.Namespaces {
			for podName, pod := range ns.Pods {
				pods = append(pods, &NamespacedPod{
					NamespaceName: nsName,
					PodName:       podName,
					Namespace:     ns,
					Pod:           pod,
				})
			}
		}
		m.allPods = &pods
	}
	return *m.allPods
}

// Namespace is the abstract representation of what matters to network policy
// tests for a namespace; i.e. it ignores kube implementation details
type Namespace struct {
	Pods   map[string]*Pod
	Labels map[string]string
}

// Pod is the abstract representation of what matters to network policy tests for
// a pod; i.e. it ignores kube implementation details
type Pod struct {
	Labels map[string]string
	IP     string
}

type NamespacedPod struct {
	NamespaceName string
	PodName       string
	Namespace     *Namespace
	Pod           *Pod
}

func (np *NamespacedPod) PodString() PodString {
	return NewPodString(np.NamespaceName, np.PodName)
}
