package kube

import (
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
)

type Resources struct {
	Namespaces map[string]map[string]string
	Pods       []*Pod
}

func NewDefaultResources(namespaces []string, podNames []string, ports []int, protocols []v1.Protocol) *Resources {
	r := &Resources{
		Namespaces: map[string]map[string]string{},
	}

	for _, ns := range namespaces {
		for _, podName := range podNames {
			r.Pods = append(r.Pods, NewPod(ns, podName, map[string]string{"pod": podName}, ports, protocols))
		}
		r.Namespaces[ns] = map[string]string{"ns": ns}
	}

	return r
}

func (r *Resources) GetJobs(port int, protocol v1.Protocol) []*Job {
	var jobs []*Job
	for _, podFrom := range r.Pods {
		for _, podTo := range r.Pods {
			jobs = append(jobs, &Job{
				FromPod:  podFrom,
				ToPod:    podTo,
				Port:     port,
				Protocol: protocol,
			})
		}
	}
	return jobs
}

func (r *Resources) NamespacesSlice() []string {
	var nss []string
	for ns := range r.Namespaces {
		nss = append(nss, ns)
	}
	return nss
}

func (r *Resources) NewTruthTable() *utils.TruthTable {
	var podNames []string
	for _, pod := range r.Pods {
		podNames = append(podNames, pod.PodString.String())
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
		_, err := kube.CreatePodIfNotExists(pod.KubePod)
		if err != nil {
			return err
		}
		_, err = kube.CreateServiceIfNotExists(pod.KubeService)
		if err != nil {
			return err
		}
	}
	return nil
}
