package connectivity

import (
	"github.com/mattfenwick/cyclonus/pkg/connectivity/kube"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/synthetic"
)

type TestCasePrinter struct {
	Noisy          bool
	IgnoreLoopback bool
}

func (t *TestCasePrinter) Print(result *TestCaseResult) {

}

type TestCaseResult struct {
	TestCase        *TestCase
	SyntheticResult *synthetic.Result
	KubeResult      *kube.Results
	Err             error // TODO how does this overlap/conflict with the err in KubeResult?
}
