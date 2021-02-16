package probe

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
)

type Probe struct {
	Ingress    *Table
	Egress     *Table
	Combined   *Table
	JobResults []*JobResult
}

func NewProbeFromJobResults(resources *Resources, jobResults []*JobResult) *Probe {
	probe := &Probe{Ingress: resources.NewTable(), Egress: resources.NewTable(), Combined: resources.NewTable(), JobResults: jobResults}
	for _, result := range jobResults {
		fr := result.Job.FromKey
		to := result.Job.ToKey
		key := fmt.Sprintf("%s/%d", result.Job.Protocol, result.Job.ResolvedPort)
		if result.Ingress != nil {
			probe.Ingress.Set(fr, to, key, *result.Ingress)
		}
		if result.Egress != nil {
			probe.Egress.Set(fr, to, key, *result.Egress)
		}
		probe.Combined.Set(fr, to, key, result.Combined)
	}
	return probe
}

func (p *Probe) ResultsByProtocol() map[v1.Protocol]map[Connectivity]int {
	byProtocol := map[v1.Protocol]map[Connectivity]int{
		v1.ProtocolTCP:  {},
		v1.ProtocolSCTP: {},
		v1.ProtocolUDP:  {},
	}
	for _, result := range p.JobResults {
		byProtocol[result.Job.Protocol][result.Combined]++
	}
	return byProtocol
}
