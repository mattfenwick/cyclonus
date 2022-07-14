package kube

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestModel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunIPAddressTests()
	RunLabelSelectorTests()
	RunSpecs(t, "network policy matcher suite")
}
