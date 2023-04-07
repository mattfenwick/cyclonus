package kube

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func RunReadNetworkPolicyTests() {
	Describe("ReadNetworkPolicies", func() {
		It("Should read a single policy from a single file", func() {
			policies, err := ReadNetworkPoliciesFromPath("../../networkpolicies/features/portrange1.yaml")
			Expect(err).To(BeNil())
			Expect(len(policies)).To(Equal(1))
		})
		It("Should read a list of policies from a single file", func() {
			policies, err := ReadNetworkPoliciesFromPath("../../networkpolicies/yaml-syntax/yaml-list.yaml")
			Expect(err).To(BeNil())
			Expect(len(policies)).To(Equal(3))
		})
		It("Should read multiple policies separated by '---' lines from a single file", func() {
			policies, err := ReadNetworkPoliciesFromPath("../../networkpolicies/yaml-syntax/triple-dash-separated.yaml")
			Expect(err).To(BeNil())
			Expect(len(policies)).To(Equal(3))
		})
		It("Should read multiple policies from all files in a directory", func() {
			policies, err := ReadNetworkPoliciesFromPath("../../networkpolicies/yaml-syntax")
			Expect(err).To(BeNil())
			Expect(len(policies)).To(Equal(6))
		})
	})
}
