package kube

import (
	"github.com/mattfenwick/cyclonus/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

type Request struct {
	Resources       *Resources
	Port            int
	Protocol        v1.Protocol
	NumberOfWorkers int
}

type Results struct {
	Request *Request
	Results []*JobResults
}

func (k *Results) TruthTable() *utils.TruthTable {
	reachability := k.Request.Resources.NewTruthTable()
	for _, result := range k.Results {
		job := result.Job
		reachability.Set(job.FromPod.PodString().String(), job.ToPod.PodString().String(), result.IsConnected)
	}
	return reachability
}
