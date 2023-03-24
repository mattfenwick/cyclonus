package connectivity

import (
	"github.com/mattfenwick/cyclonus/pkg/connectivity/probe"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/sirupsen/logrus"
	networkingv1 "k8s.io/api/networking/v1"
)

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
		logrus.Debugf("comparing i [%d] in kubeprobes %+v", i, s.KubeProbes)
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

func (s *StepResult) Passed(ignoreLoopback bool) bool {
	return s.LastComparison().ValueCounts(ignoreLoopback)[DifferentComparison] == 0
}
