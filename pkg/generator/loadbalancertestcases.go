package generator

import (
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (t *TestCaseGenerator) LoadBalancerTestCase() []*TestCase {
	svc1 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service-name",
			Namespace: "x",
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Ports: []v1.ServicePort{
				{
					Protocol: v1.ProtocolTCP,
					Port:     81,
					NodePort: 32086,
				},
			},
			Selector: map[string]string{"pod": "a"},
		},
	}
	probe := &ProbeConfig{
		AllAvailable: false,
		PortProtocol: &PortProtocol{
			Protocol: v1.ProtocolTCP,
			Port:     intstr.FromInt(32086),
		},
		Mode: ProbeModeNodeIP,
	}
	denyAll := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deny-all-policy",
			Namespace: "x",
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
		},
	}

	return []*TestCase{
		NewTestCase("should allow access to nodeport with no netpols applied",
			NewStringSet(TagLoadBalancer),
			NewTestStep(probe,
				CreateService(svc1),
			),
		),
		NewTestCase("should deny access to nodeport with netpols applied",
			NewStringSet(TagLoadBalancer),
			NewTestStep(probe,
				CreatePolicy(denyAll),
			),
		),
	}
}
