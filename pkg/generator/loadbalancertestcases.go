package generator

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (t *TestCaseGenerator) LoadBalancerTestCase() []*TestCase {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service-name",
			Namespace: "x",
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Ports: []v1.ServicePort{
				{
					Protocol: v1.ProtocolTCP,
					Port:     8087,
					NodePort: 32087,
				},
			},
			Selector: map[string]string{"app": "toolbox"},
		},
	}
	probe := &ProbeConfig{
		AllAvailable: false,
		PortProtocol: &PortProtocol{
			Protocol: v1.ProtocolTCP,
			Port:     intstr.FromInt(32087),
		},
		Mode: ProbeModeNodeIP,
	}
	return []*TestCase{
		NewTestCase("should allow ingress access on one named port",
			NewStringSet(TagLoadBalancer),
			NewTestStep(probe, CreateService(svc))),
	}
}
