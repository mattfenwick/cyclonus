package generator

import (
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	portServe81TCP = &ProbeConfig{PortProtocol: &matcher.PortProtocol{Port: intstr.FromString("serve-81-tcp"), Protocol: tcp}}
)

type ExampleGenerator struct{}

func (e *ExampleGenerator) GenerateTestCases() []*TestCase {
	return []*TestCase{
		NewTestCase("should allow ingress access on one named port",
			NewTestStep(portServe81TCP, CreatePolicy(
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
			NewTestStep(port80TCP)),
	}
}
