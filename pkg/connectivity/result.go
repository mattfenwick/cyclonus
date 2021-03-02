package connectivity

import (
	"fmt"
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

func (r *Result) Features() ([]string, []string, []string, []string) {
	return r.TestCase.GetFeatures()
}

type CombinedResults struct {
	Results []*Result
}

type Summary struct {
	Tests          [][]string
	Passed         int
	Failed         int
	ProtocolCounts map[v1.Protocol]map[Comparison]int
	TagCounts      map[string]map[bool]map[string]int
	//FeatureCounts map[string]map[bool]map[string]int
}

func (c *CombinedResults) Summary(ignoreLoopback bool) *Summary {
	summary := &Summary{
		Tests:          nil,
		Passed:         0,
		Failed:         0,
		ProtocolCounts: map[v1.Protocol]map[Comparison]int{v1.ProtocolTCP: {}, v1.ProtocolSCTP: {}, v1.ProtocolUDP: {}},
		TagCounts:      map[string]map[bool]map[string]int{},
	}
	passedTotal, failedTotal := 0, 0
	// TODO restore these
	//generalPassFailCounts := map[bool]map[string]int{false: {}, true: {}}
	//ingressPassFailCounts := map[bool]map[string]int{false: {}, true: {}}
	//egressPassFailCounts := map[bool]map[string]int{false: {}, true: {}}
	//actionPassFailCounts := map[bool]map[string]int{false: {}, true: {}}

	for testNumber, result := range c.Results {
		// preprocess to figure out whether it passed or failed
		passed := true
		for _, step := range result.Steps {
			if step.LastComparison().ValueCounts(ignoreLoopback)[DifferentComparison] > 0 {
				passed = false
			}
		}

		//general, ingress, egress, actions := result.Features()
		//incrementCounts(generalPassFailCounts, passed, general)
		//incrementCounts(ingressPassFailCounts, passed, ingress)
		//incrementCounts(egressPassFailCounts, passed, egress)
		//incrementCounts(actionPassFailCounts, passed, actions)
		groupedTags := result.TestCase.Tags.GroupTags()
		for primary, subs := range groupedTags {
			if _, ok := summary.TagCounts[primary]; !ok {
				summary.TagCounts[primary] = map[bool]map[string]int{true: {}, false: {}}
			}
			incrementCounts(summary.TagCounts[primary], passed, subs)
		}

		var testResult string
		if passed {
			testResult = "passed"
			passedTotal++
		} else {
			testResult = "failed"
			failedTotal++
		}

		summary.Tests = append(summary.Tests, []string{
			fmt.Sprintf("%d: %s", testNumber+1, result.TestCase.Description),
			testResult, "", "", "", "",
			"", "", "",
		})

		for stepNumber, step := range result.Steps {
			for tryNumber := range step.KubeProbes {
				counts := step.Comparison(tryNumber).ValueCounts(ignoreLoopback)
				tryProtocolCounts := step.Comparison(tryNumber).ValueCountsByProtocol(ignoreLoopback)
				tcp := tryProtocolCounts[v1.ProtocolTCP]
				sctp := tryProtocolCounts[v1.ProtocolSCTP]
				udp := tryProtocolCounts[v1.ProtocolUDP]
				summary.Tests = append(summary.Tests, []string{
					"",
					"",
					fmt.Sprintf("Step %d, try %d", stepNumber+1, tryNumber+1),
					intToString(counts[DifferentComparison]),
					intToString(counts[SameComparison]),
					intToString(counts[IgnoredComparison]),
					protocolResult(tcp[SameComparison], tcp[DifferentComparison]),
					protocolResult(sctp[SameComparison], sctp[DifferentComparison]),
					protocolResult(udp[SameComparison], udp[DifferentComparison]),
				})

				summary.ProtocolCounts[v1.ProtocolTCP][SameComparison] += tcp[SameComparison]
				summary.ProtocolCounts[v1.ProtocolTCP][DifferentComparison] += tcp[DifferentComparison]
				summary.ProtocolCounts[v1.ProtocolSCTP][SameComparison] += sctp[SameComparison]
				summary.ProtocolCounts[v1.ProtocolSCTP][DifferentComparison] += sctp[DifferentComparison]
				summary.ProtocolCounts[v1.ProtocolUDP][SameComparison] += udp[SameComparison]
				summary.ProtocolCounts[v1.ProtocolUDP][DifferentComparison] += udp[DifferentComparison]
			}
		}
	}

	summary.Passed = passedTotal
	summary.Failed = failedTotal

	//fmt.Println(passFailTable("general", generalPassFailCounts, &passedTotal, &failedTotal))
	//fmt.Println(passFailTable("ingress", ingressPassFailCounts, nil, nil))
	//fmt.Println(passFailTable("egress", egressPassFailCounts, nil, nil))
	//fmt.Println(passFailTable("actions", actionPassFailCounts, nil, nil))

	return summary
}
