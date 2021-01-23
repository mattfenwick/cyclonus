package kube

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strings"
)

type Resources struct {
	Namespaces map[string]map[string]string
	Pods       []*Pod
	Jobs       []*Job
}

func NewDefaultResources(namespaces []string, podNames []string, port int, protocol v1.Protocol) *Resources {
	r := &Resources{
		Namespaces: map[string]map[string]string{},
	}

	for _, ns := range namespaces {
		for _, podName := range podNames {
			r.Pods = append(r.Pods, &Pod{
				Namespace:     ns,
				Name:          podName,
				Labels:        map[string]string{"pod": podName},
				ContainerName: fmt.Sprintf("cont-%d-%s", port, strings.ToLower(string(protocol))),
				Port:          port,
				Protocol:      protocol,
			})
		}
		r.Namespaces[ns] = map[string]string{"ns": ns}
	}

	for _, podFrom := range r.Pods {
		for _, podTo := range r.Pods {
			r.Jobs = append(r.Jobs, &Job{
				FromPod: podFrom,
				ToPod:   podTo,
			})
		}
	}
	return r
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
