package connectivity

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"sort"
	"strings"
)

type PodModel struct {
	Namespaces map[string]*Namespace
	// derived
	allPodStrings *[]PodString
	allPods       *[]*NamespacedPod
}

func NewDefaultModel(namespaces []string, podNames []string, port int, protocol v1.Protocol) *PodModel {
	model := &PodModel{Namespaces: map[string]*Namespace{}}

	for _, ns := range namespaces {
		pods := map[string]*Pod{}
		for _, podName := range podNames {
			pods[podName] = &Pod{
				Labels: map[string]string{"pod": podName},
				Containers: []*Container{
					{
						Name:     fmt.Sprintf("cont-%d-%s", port, strings.ToLower(string(protocol))),
						Port:     port,
						Protocol: protocol,
					},
				},
			}
		}
		model.Namespaces[ns] = &Namespace{Pods: pods, Labels: map[string]string{"ns": ns}}
	}
	return model
}

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
					Containers:    pod.Containers,
				})
			}
		}
		m.allPods = &pods
	}
	return *m.allPods
}

type Namespace struct {
	Pods   map[string]*Pod
	Labels map[string]string
}

type Pod struct {
	Labels     map[string]string
	IP         string
	Containers []*Container
}

type Container struct {
	Name     string
	Port     int
	Protocol v1.Protocol
}

type NamespacedPod struct {
	NamespaceName string
	PodName       string
	Namespace     *Namespace
	Pod           *Pod
	Containers    []*Container
}

func (np *NamespacedPod) PodString() PodString {
	return NewPodString(np.NamespaceName, np.PodName)
}

func ServiceName(namespace string, pod string) string {
	return fmt.Sprintf("s-%s-%s", namespace, pod)
}

func (np *NamespacedPod) ServiceName() string {
	return ServiceName(np.NamespaceName, np.PodName)
}
