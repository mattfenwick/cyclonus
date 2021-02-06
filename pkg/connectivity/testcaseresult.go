package connectivity

import (
	"github.com/mattfenwick/cyclonus/pkg/connectivity/types"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	networkingv1 "k8s.io/api/networking/v1"
)

type Result struct {
	TestCase *generator.TestCase
	Steps    []*StepResult
	Err      error
}

type StepResult struct {
	SimulatedProbe *types.Probe
	KubeProbes     []*types.Table
	Policy         *matcher.Policy
	KubePolicies   []*networkingv1.NetworkPolicy
}

func (s *StepResult) LastKubeProbe() *types.Table {
	return s.KubeProbes[len(s.KubeProbes)-1]
}
