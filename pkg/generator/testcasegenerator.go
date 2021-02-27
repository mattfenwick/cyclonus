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

func (t *TestCaseGeneratorReplacement) GenerateAllTestCases() []*TestCase {
	return t.PortProtocolTestCases()
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
