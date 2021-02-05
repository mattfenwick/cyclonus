package kube

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Request struct {
	Resources       *Resources
	Port            intstr.IntOrString
	Protocol        v1.Protocol
	NumberOfWorkers int
}

type Results struct {
	Request *Request
	Results []*JobResult
}

func (k *Results) TruthTable() *ResultTable {
	reachability := k.Request.Resources.NewResultTable()
	for _, result := range k.Results {
		job := result.Job
		reachability.Set(job.FromPod.PodString.String(), job.ToPod.PodString.String(), result.Connectivity)
	}
	return reachability
}
