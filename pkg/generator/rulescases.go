package generator

func (t *TestCaseGeneratorReplacement) RulesTestCases() []*TestCase {
	// TODO break rules into length 0, 1, 2, etc.
	var cases []*TestCase
	for _, isIngress := range []bool{false, true} {
		direction := describeDirectionality(isIngress)
		cases = append(cases, NewSingleStepTestCase("", NewStringSet(direction, TagDenyAll), ProbeAllAvailable,
			CreatePolicy(BuildPolicy(SetRules(isIngress, DenyAllRules)).NetworkPolicy())))
		cases = append(cases, NewSingleStepTestCase("", NewStringSet(direction, TagAllowAll), ProbeAllAvailable,
			CreatePolicy(BuildPolicy(SetRules(isIngress, AllowAllRules)).NetworkPolicy())))
	}
	return cases
}
