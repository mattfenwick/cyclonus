package connectivity

import (
	"github.com/mattfenwick/cyclonus/pkg/connectivity/probe"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
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
	featureMap := map[string]bool{}
	//for _, step := range r.TestCase.Steps {
	//	if step.Probe.AllAvailable {
	//		for protocol := range r.InitialResources.AllProtocolsServed() {
	//			featureMap[generator.ProtocolToFeature(protocol)] = true
	//		}
	//	} else if step.Probe.PortProtocol != nil {
	//		pp := step.Probe.PortProtocol
	//		featureMap[generator.ProtocolToFeature(pp.Protocol)] = true
	//		switch pp.Port.Type {
	//		case intstr.Int:
	//			featureMap[generator.ProbeFeatureNumberedPort] = true
	//		case intstr.String:
	//			featureMap[generator.ProbeFeatureNamedPort] = true
	//		default:
	//			panic(errors.Errorf("invalid intstr value %T", pp.Port))
	//		}
	//	} else {
	//		panic(errors.Errorf("invalid ProbeConfig value %T", step.Probe))
	//	}
	//}
	var features []string
	for feature := range featureMap {
		features = append(features, feature)
	}
	return features
}

func (r *Result) Features() []string {
	return append(r.TestCase.GetFeatures().Strings(), r.ProbeFeatures()...)
}

type StepResult struct {
	SimulatedProbe *probe.Probe
	KubeProbes     []*probe.Table
	Policy         *matcher.Policy
	KubePolicies   []*networkingv1.NetworkPolicy
}

func (s *StepResult) LastKubeProbe() *probe.Table {
	return s.KubeProbes[len(s.KubeProbes)-1]
}
