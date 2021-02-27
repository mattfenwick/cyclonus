package generator

type TestCaseGenerator interface {
	GenerateTestCases() []*TestCase
}

type TestCaseGeneratorReplacement struct {
	PodIP    string
	AllowDNS bool
	Tags     []string
}

func NewTestCaseGeneratorReplacement(allowDNS bool, podIP string, tags []string) *TestCaseGeneratorReplacement {
	return &TestCaseGeneratorReplacement{
		PodIP:    podIP,
		AllowDNS: allowDNS,
		Tags:     tags,
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
		t.PortProtocolTestCases(),
		t.ExampleTestCases(),
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
