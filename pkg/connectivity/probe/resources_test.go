package probe

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
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

		It("Should set pod labels nondestructively", func() {
			labels := map[string]string{"pod": "b"}
			r := &Resources{
				Namespaces: map[string]map[string]string{
					"y": {},
				},
				Pods: []*Pod{{Namespace: "y", Name: "b", Labels: labels}},
			}
			r2, err := r.SetPodLabels("y", "b", map[string]string{})
			Expect(err).To(Succeed())

			Expect(r.Pods[0].Labels).To(Equal(labels))
			Expect(r2.Pods[0].Labels).To(Equal(map[string]string{}))
		})

		It("Should create a service nondestructively", func() {
			r := &Resources{
				Services: make(map[string]*v1.Service),
			}
			svc := v1.Service{}
			r2, err := r.CreateService(&svc)
			Expect(err).To(Succeed())

			Expect(r.Services).To(HaveLen(0))
			Expect(r2.Services).To(HaveLen(1))
		})
	})
}
