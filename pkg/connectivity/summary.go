package connectivity

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
)

type SummaryTable struct {
	Tests                [][]string
	Passed               int
	Failed               int
	ProtocolCounts       map[v1.Protocol]map[Comparison]int
	TagCounts            map[string]map[string]map[bool]int
	TagPrimaryCounts     map[string]map[bool]int
	FeatureCounts        map[string]map[string]map[bool]int
	FeaturePrimaryCounts map[string]map[bool]int
}

func NewSummaryTableFromResults(ignoreLoopback bool, results []*Result) *SummaryTable {
	summary := &SummaryTable{
		Tests:                nil,
		Passed:               0,
		Failed:               0,
		ProtocolCounts:       map[v1.Protocol]map[Comparison]int{v1.ProtocolTCP: {}, v1.ProtocolSCTP: {}, v1.ProtocolUDP: {}},
		TagCounts:            map[string]map[string]map[bool]int{},
		TagPrimaryCounts:     map[string]map[bool]int{},
		FeatureCounts:        map[string]map[string]map[bool]int{},
		FeaturePrimaryCounts: map[string]map[bool]int{},
	}
	passedTotal, failedTotal := 0, 0

	for testNumber, result := range results {
		passed := result.Passed(ignoreLoopback)

		for primary, subs := range result.Features() {
			if _, ok := summary.FeatureCounts[primary]; !ok {
				summary.FeatureCounts[primary] = map[string]map[bool]int{}
			}
			incrementCounts(summary.FeatureCounts[primary], subs, passed)
			incrementCounts(summary.FeaturePrimaryCounts, []string{primary}, passed)
		}

		groupedTags := result.TestCase.Tags.GroupTags()
		for primary, subs := range groupedTags {
			if _, ok := summary.TagCounts[primary]; !ok {
				summary.TagCounts[primary] = map[string]map[bool]int{}
			}
			incrementCounts(summary.TagCounts[primary], subs, passed)
			incrementCounts(summary.TagPrimaryCounts, []string{primary}, passed)
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

	return summary
}

func incrementCounts(dict map[string]map[bool]int, keys []string, b bool) {
	for _, k := range keys {
		if _, ok := dict[k]; !ok {
			dict[k] = map[bool]int{}
		}
		dict[k][b]++
	}
}

func protocolResult(passed int, failed int) string {
	total := passed + failed
	if total == 0 {
		return "-"
	}
	return fmt.Sprintf("%d / %d (%.0f%%)", passed, total, percentage(passed, total))
}
