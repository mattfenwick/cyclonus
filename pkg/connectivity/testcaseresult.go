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

func (r *Result) ProbeFeatures() []string {
	featureMap := map[v1.Protocol]bool{}
	for _, step := range r.Steps {
		for _, counts := range NewComparisonTableFrom(step.LastKubeProbe(), step.SimulatedProbe).ResultsByProtocol() {
			for protocol, count := range counts {
				if count > 0 {
					featureMap[protocol] = true
				}
			}
		}
	}
	var features []string
	for feature := range featureMap {
		features = append(features, probe.ProtocolToFeature(feature))
	}
	return features
}

func (r *Result) Features() []string {
	return append(r.TestCase.GetFeatures().Strings(), r.ProbeFeatures()...)
}

type StepResult struct {
	SimulatedProbe *probe.Table
	KubeProbes     []*probe.Table
	Policy         *matcher.Policy
	KubePolicies   []*networkingv1.NetworkPolicy
}

func (s *StepResult) LastKubeProbe() *probe.Table {
	return s.KubeProbes[len(s.KubeProbes)-1]
}
