package generator

import (
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (t *TestCaseGenerator) ExampleTestCases() []*TestCase {
	return []*TestCase{
		NewTestCase("should allow ingress access on one named port",
			NewStringSet(TagExample),
			NewTestStep(ProbeAllAvailable, CreatePolicy(
				&NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "allow-all",
						Namespace: "x",
					},
					Spec: NetworkPolicySpec{
						PodSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{},
						},
						Ingress: []NetworkPolicyIngressRule{
							{
								Ports: []NetworkPolicyPort{
									{Port: &intstr.IntOrString{Type: intstr.String, StrVal: "serve-81-tcp"}},
								},
							},
						},
						PolicyTypes: []PolicyType{PolicyTypeIngress},
					},
				})),
			NewTestStep(ProbeAllAvailable,
				CreateNamespace("w", map[string]string{"ns": "w"}),
				CreatePod("w", "a", map[string]string{"pod": "a"})),
			NewTestStep(ProbeAllAvailable, DeletePod("w", "a")),
			NewTestStep(ProbeAllAvailable, DeleteNamespace("w")),
			NewTestStep(ProbeAllAvailable),
			//NewTestStep(probePort80TCP),
			NewTestStep(probePort81TCP),
			//NewTestStep(probePortServe80TCP),
			NewTestStep(probePortServe81TCP)),
	}
}
