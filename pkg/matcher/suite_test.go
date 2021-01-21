package matcher

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMatcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunBuilderTests()
	RunPolicyTests()
	RunSpecs(t, "network policy matcher suite")
}
