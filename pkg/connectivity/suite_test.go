package connectivity

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConnectivity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunTestCaseStateTests()
	RunPrinterTests()
	RunSpecs(t, "connectivity suite")
}
