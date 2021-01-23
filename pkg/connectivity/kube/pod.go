package kube

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const (
	agnhostImage = "k8s.gcr.io/e2e-test-images/agnhost:2.21"
)

type Pod struct {
	Namespace     string
	Name          string
	Labels        map[string]string
	ContainerName string
	Port          int
	Protocol      v1.Protocol
}

func (p *Pod) ServiceName() string {
	return fmt.Sprintf("s-%s-%s", p.Namespace, p.Name)
}

func (p *Pod) PodString() utils.PodString {
	return utils.NewPodString(p.Namespace, p.Name)
}

// KubePod returns the kube pod
func (p *Pod) KubePod() *v1.Pod {
	zero := int64(0)
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name,
			Labels:    p.Labels,
			Namespace: p.Namespace,
		},
		Spec: v1.PodSpec{
			TerminationGracePeriodSeconds: &zero,
			Containers:                    []v1.Container{p.KubeContainer()},
		},
	}
}

// Service returns a kube service spec
func (p *Pod) KubeService() *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.ServiceName(),
			Namespace: p.Namespace,
		},
		Spec: v1.ServiceSpec{
			Selector: p.Labels,
		},
	}

	service.Spec.Ports = append(service.Spec.Ports, v1.ServicePort{
		Name:     fmt.Sprintf("service-port-%s-%d", strings.ToLower(string(p.Protocol)), p.Port),
		Protocol: p.Protocol,
		Port:     int32(p.Port),
	})

	return service
}

func (p *Pod) KubeContainer() v1.Container {
	var cmd []string

	port := p.Port
	switch p.Protocol {
	case v1.ProtocolTCP:
		cmd = []string{"/agnhost", "serve-hostname", "--tcp", "--http=false", "--port", fmt.Sprintf("%d", port)}
	case v1.ProtocolUDP:
		cmd = []string{"/agnhost", "serve-hostname", "--udp", "--http=false", "--port", fmt.Sprintf("%d", port)}
	case v1.ProtocolSCTP:
		cmd = []string{"/agnhost", "netexec", "--sctp-port", fmt.Sprintf("%d", port)}
	default:
		panic(errors.Errorf("invalid protocol %s", p.Protocol))
	}
	return v1.Container{
		Name:            p.ContainerName,
		ImagePullPolicy: v1.PullIfNotPresent,
		Image:           agnhostImage,
		Command:         cmd,
		SecurityContext: &v1.SecurityContext{},
		Ports: []v1.ContainerPort{
			{
				ContainerPort: int32(port),
				Name:          fmt.Sprintf("serve-%d-%s", port, strings.ToLower(string(p.Protocol))),
				Protocol:      p.Protocol,
			},
		},
	}
}
