package connectivity

import "github.com/mattfenwick/cyclonus/pkg/connectivity/kube"

type TestCasePrinter struct {
	Noisy          bool
	IgnoreLoopback bool
}

func (t *TestCasePrinter) Print(result *TestCaseResult) {

}

type TestCaseResult struct {
	TestCase        *TestCase
	SyntheticResult *SyntheticProbeResult
	KubeResult      *kube.Results
	Err             error // TODO how does this overlap/conflict with the err in KubeResult?
}
