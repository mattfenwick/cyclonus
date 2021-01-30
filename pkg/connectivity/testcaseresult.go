package connectivity

import (
	connectivitykube "github.com/mattfenwick/cyclonus/pkg/connectivity/kube"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/synthetic"
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
	SyntheticResult *synthetic.Result
	KubeResults     []*connectivitykube.Results
	Policy          *matcher.Policy
	KubePolicies    []*networkingv1.NetworkPolicy
}

func (s *StepResult) LastKubeResult() *connectivitykube.Results {
	return s.KubeResults[len(s.KubeResults)-1]
}
