package generator

type TestCaseGenerator interface {
	GenerateTestCases() []*TestCase
}

/*
TODO
Test cases:

1 policy with ingress:
 - empty ingress
 - ingress with 1 rule
   - empty
   - 1 port
     - empty
     - protocol
     - port
     - port + protocol
   - 2 ports
   - 1 from
     - 8 combos: (nil + nil => might mean ipblock must be non-nil)
       - pod sel: nil, empty, non-empty
       - ns sel: nil, empty, non-empty
     - ipblock
       - no except
       - yes except
   - 2 froms
     - 1 pod/ns, 1 ipblock
     - 2 pod/ns
     - 2 ipblocks
   - 1 port, 1 from
   - 2 ports, 2 froms
 - ingress with 2 rules
 - ingress with 3 rules
2 policies with ingress
1 policy with egress
2 policies with egress
1 policy with both ingress and egress
2 policies with both ingress and egress
*/
type TestCaseGeneratorReplacement struct {
	PodIP      string
	AllowDNS   bool
	Tags       []string
	Namespaces []string
}

func NewTestCaseGeneratorReplacement(allowDNS bool, podIP string, tags []string, namespaces []string) *TestCaseGeneratorReplacement {
	return &TestCaseGeneratorReplacement{
		PodIP:      podIP,
		AllowDNS:   allowDNS,
		Tags:       tags,
		Namespaces: namespaces,
	}
}

func flatten(caseSlices ...[]*TestCase) []*TestCase {
	var cases []*TestCase
	for _, slice := range caseSlices {
		cases = append(cases, slice...)
	}
	return cases
}

func (t *TestCaseGeneratorReplacement) GenerateAllTestCases() []*TestCase {
	return flatten(
		t.TargetTestCases(),
		t.RulesTestCases(),
		t.PeersTestCases(),
		t.PortProtocolTestCases(),
		t.ExampleTestCases(),
		t.ActionTestCases(),
		t.UpstreamE2ETestCases())
}

func (t *TestCaseGeneratorReplacement) GenerateTestCases() []*TestCase {
	var cases []*TestCase
	for _, testcase := range t.GenerateAllTestCases() {
		if len(t.Tags) == 0 || testcase.Tags.ContainsAny(t.Tags) {
			cases = append(cases, testcase)
		}
	}
	return cases
}
