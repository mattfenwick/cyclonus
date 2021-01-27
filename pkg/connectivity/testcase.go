package connectivity

import (
	connectivitykube "github.com/mattfenwick/cyclonus/pkg/connectivity/kube"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/synthetic"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

type TestCase struct {
	KubePolicies              []*networkingv1.NetworkPolicy
	NetpolCreationWaitSeconds int
	Port                      int
	Protocol                  v1.Protocol
	KubeResources             *connectivitykube.Resources
	SyntheticResources        *synthetic.Resources
}

type TestCaseResult struct {
	TestCase        *TestCase
	SyntheticResult *synthetic.Result
	KubeResult      *connectivitykube.Results
	Policy          *matcher.Policy
	Err             error // TODO how does this overlap/conflict with the err in KubeResult?
}
