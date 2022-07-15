package matcher

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMatcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunBuilderTests()
	RunPolicyTests()
	RunSimplifierTests()
	RunSpecs(t, "network policy matcher suite")
}
