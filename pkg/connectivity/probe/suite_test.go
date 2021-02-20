package probe

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestProbe(t *testing.T) {
	RegisterFailHandler(Fail)
	RunResourcesTests()
	RunSpecs(t, "generator suite")
}
