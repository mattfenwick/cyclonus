package matcher

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestModel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunBuilderTests()
	RunExplainerTests()
	RunPolicyTests()
	RunSpecs(t, "network policy matcher suite")
}
