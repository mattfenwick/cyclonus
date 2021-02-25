package connectivity

import (
	"github.com/mattfenwick/cyclonus/pkg/connectivity/probe"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
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

func (r *Result) Features() ([]string, []string, []string, []string) {
	return r.TestCase.GetFeatures()
}

type StepResult struct {
	SimulatedProbe *probe.Table
	KubeProbes     []*probe.Table
	Policy         *matcher.Policy
	KubePolicies   []*networkingv1.NetworkPolicy
	comparisons    []*ComparisonTable
}

func NewStepResult(simulated *probe.Table, policy *matcher.Policy, kubePolicies []*networkingv1.NetworkPolicy) *StepResult {
	return &StepResult{
		SimulatedProbe: simulated,
		Policy:         policy,
		KubePolicies:   kubePolicies,
	}
}

func (s *StepResult) AddKubeProbe(kubeProbe *probe.Table) {
	s.KubeProbes = append(s.KubeProbes, kubeProbe)
	s.comparisons = append(s.comparisons, nil)
}

func (s *StepResult) Comparison(i int) *ComparisonTable {
	if s.comparisons[i] == nil {
		s.comparisons[i] = NewComparisonTableFrom(s.KubeProbes[i], s.SimulatedProbe)
	}
	return s.comparisons[i]
}

func (s *StepResult) LastComparison() *ComparisonTable {
	return s.Comparison(len(s.KubeProbes) - 1)
}

func (s *StepResult) LastKubeProbe() *probe.Table {
	return s.KubeProbes[len(s.KubeProbes)-1]
}
