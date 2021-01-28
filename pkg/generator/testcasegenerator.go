package generator

type TestCaseGenerator interface {
	GenerateTestCases() []*TestCase
}
