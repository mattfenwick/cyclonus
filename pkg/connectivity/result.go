package connectivity

import (
	"github.com/mattfenwick/cyclonus/pkg/connectivity/probe"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	v1 "k8s.io/api/core/v1"
)

type Result struct {
	// TODO should resources be captured per-step for tests that modify them?
	InitialResources *probe.Resources
	TestCase         *generator.TestCase
	Steps            []*StepResult
	Err              error
}

func (r *Result) ResultsByProtocol() map[bool]map[v1.Protocol]int {
	counts := map[bool]map[v1.Protocol]int{true: {}, false: {}}
	for _, step := range r.Steps {
		for isSuccess, protocolCounts := range step.LastComparison().ResultsByProtocol() {
			for protocol, count := range protocolCounts {
				counts[isSuccess][protocol] += count
			}
		}
	}
	return counts
}

func (r *Result) Features() map[string][]string {
	return r.TestCase.GetFeatures()
}

func (r *Result) Passed(ignoreLoopback bool) bool {
	for _, step := range r.Steps {
		if step.LastComparison().ValueCounts(ignoreLoopback)[DifferentComparison] > 0 {
			return false
		}
	}
	return true
}
