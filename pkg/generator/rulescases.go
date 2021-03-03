package generator

import "fmt"

func (t *TestCaseGenerator) RulesTestCases() []*TestCase {
	// TODO break rules into length 0, 1, 2, etc.
	var cases []*TestCase
	for _, isIngress := range []bool{false, true} {
		dir := describeDirectionality(isIngress)
		cases = append(cases, NewSingleStepTestCase(fmt.Sprintf("%s: deny all", dir), NewStringSet(dir, TagDenyAll), ProbeAllAvailable,
			CreatePolicy(BuildPolicy(SetRules(isIngress, DenyAllRules)).NetworkPolicy())))
		cases = append(cases, NewSingleStepTestCase(fmt.Sprintf("%s: allow all", dir), NewStringSet(dir, TagAllowAll), ProbeAllAvailable,
			CreatePolicy(BuildPolicy(SetRules(isIngress, AllowAllRules)).NetworkPolicy())))
	}
	return cases
}
