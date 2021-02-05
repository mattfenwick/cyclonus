package kube

import (
	"github.com/mattfenwick/cyclonus/pkg/kube"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sort"
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
			r.Pods = append(r.Pods, NewPod(ns, podName, map[string]string{"pod": podName}, ports, protocols))
		}
		r.Namespaces[ns] = map[string]string{"ns": ns}
	}

	return r
}

type Jobs struct {
	Valid           []*Job
	BadNamedPort    []*Job
	BadPortProtocol []*Job
}

func (r *Resources) GetJobsForSpecificPortProtocol(port intstr.IntOrString, protocol v1.Protocol) *Jobs {
	jobs := &Jobs{}
	for _, podFrom := range r.Pods {
		for _, podTo := range r.Pods {
			var portInt int
			var err error
			switch port.Type {
			case intstr.Int:
				portInt = int(port.IntVal)
			case intstr.String:
				portInt, err = podTo.ResolveNamedPort(port.StrVal)
			}
			if err != nil {
				jobs.BadNamedPort = append(jobs.BadNamedPort, &Job{
					FromPod:  podFrom,
					ToPod:    podTo,
					Port:     -1,
					Protocol: protocol,
				})
				continue
			}
			job := &Job{
				FromPod:  podFrom,
				ToPod:    podTo,
				Port:     portInt,
				Protocol: protocol,
			}
			if !podTo.IsServingPortProtocol(portInt, protocol) {
				jobs.BadPortProtocol = append(jobs.BadPortProtocol, job)
			} else {
				jobs.Valid = append(jobs.Valid, job)
			}
		}
		// TODO from pod to external ip
		//for _, ip := range r.ExternalIPs {
		//
		//}
		// TODO no way to do from external ip to pod?
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

func (r *Resources) NewResultTable() *ResultTable {
	var podNames []string
	for _, pod := range r.Pods {
		podNames = append(podNames, pod.PodString.String())
	}
	sort.Strings(podNames)
	return NewResultTable(append(podNames, r.ExternalIPs...))
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
