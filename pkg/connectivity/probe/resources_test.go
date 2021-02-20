package probe

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func RunResourcesTests() {
	Describe("Resources", func() {
		It("Should add a namespace nondestructively", func() {
			r := &Resources{
				Namespaces: map[string]map[string]string{
					"x": {},
				},
				Pods: []*Pod{{Namespace: "x", Name: "a"}},
			}
			r2, err := r.CreateNamespace("y", map[string]string{})
			Expect(err).To(Succeed())

			Expect(r.Namespaces).To(HaveLen(1))
			Expect(r2.Namespaces).To(HaveLen(2))
		})

		It("Should add a pod nondestructively", func() {
			r := &Resources{
				Namespaces: map[string]map[string]string{
					"x": {},
				},
				Pods: []*Pod{{Namespace: "x", Name: "a"}},
			}
			r2, err := r.CreatePod("x", "b", map[string]string{})
			Expect(err).To(Succeed())

			Expect(r.Pods).To(HaveLen(1))
			Expect(r2.Pods).To(HaveLen(2))
		})
	})
}
