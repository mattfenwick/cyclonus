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
	Namespace  string
	Name       string
	Labels     map[string]string
	Containers []*Container
	// derived
	KubeService *v1.Service
	PodString   utils.PodString
	KubePod     *v1.Pod
}

func NewPod(ns string, name string, labels map[string]string, ports []int, protocols []v1.Protocol) *Pod {
	p := &Pod{
		Namespace: ns,
		Name:      name,
		Labels:    labels,
	}
	for _, port := range ports {
		for _, protocol := range protocols {
			p.Containers = append(p.Containers, &Container{Port: port, Protocol: protocol})
		}
	}
	p.KubeService = p.kubeService()
	p.PodString = p.podString()
	p.KubePod = p.kubePod()
	return p
}

func (p *Pod) ServiceName() string {
	return fmt.Sprintf("s-%s-%s", p.Namespace, p.Name)
}

func (p *Pod) podString() utils.PodString {
	return utils.NewPodString(p.Namespace, p.Name)
}

// KubePod returns the kube pod
func (p *Pod) kubePod() *v1.Pod {
	zero := int64(0)
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name,
			Labels:    p.Labels,
			Namespace: p.Namespace,
		},
		Spec: v1.PodSpec{
			TerminationGracePeriodSeconds: &zero,
			Containers:                    p.KubeContainers(),
		},
	}
}

// Service returns a kube service spec
func (p *Pod) kubeService() *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.ServiceName(),
			Namespace: p.Namespace,
		},
		Spec: v1.ServiceSpec{
			Selector: p.Labels,
		},
	}

	for _, cont := range p.Containers {
		service.Spec.Ports = append(service.Spec.Ports, cont.KubeServicePort())
	}

	return service
}

func (p *Pod) KubeContainers() []v1.Container {
	var containers []v1.Container
	for _, cont := range p.Containers {
		containers = append(containers, cont.KubeContainer())
	}
	return containers
}

func (p *Pod) ResolveNamedPort(port string) (int, error) {
	for _, c := range p.Containers {
		if c.ContainerPortName() == port {
			return c.Port, nil
		}
	}
	return 0, errors.Errorf("unable to resolve named port %s on pod %s/%s", port, p.Namespace, p.Name)
}

func (p *Pod) IsServingPortProtocol(port int, protocol v1.Protocol) bool {
	for _, cont := range p.Containers {
		if cont.Port == port && cont.Protocol == protocol {
			return true
		}
	}
	return false
}

type Container struct {
	Port     int
	Protocol v1.Protocol
}

func (c *Container) Name() string {
	return fmt.Sprintf("cont-%d-%s", c.Port, strings.ToLower(string(c.Protocol)))
}

func (c *Container) ContainerPortName() string {
	return fmt.Sprintf("serve-%d-%s", c.Port, strings.ToLower(string(c.Protocol)))
}

func (c *Container) KubeServicePort() v1.ServicePort {
	return v1.ServicePort{
		Name:     fmt.Sprintf("service-port-%s-%d", strings.ToLower(string(c.Protocol)), c.Port),
		Protocol: c.Protocol,
		Port:     int32(c.Port),
	}
}

func (c *Container) KubeContainer() v1.Container {
	var cmd []string

	switch c.Protocol {
	case v1.ProtocolTCP:
		cmd = []string{"/agnhost", "serve-hostname", "--tcp", "--http=false", "--port", fmt.Sprintf("%d", c.Port)}
	case v1.ProtocolUDP:
		cmd = []string{"/agnhost", "serve-hostname", "--udp", "--http=false", "--port", fmt.Sprintf("%d", c.Port)}
	case v1.ProtocolSCTP:
		cmd = []string{"/agnhost", "netexec", "--sctp-port", fmt.Sprintf("%d", c.Port)}
	default:
		panic(errors.Errorf("invalid protocol %s", c.Protocol))
	}
	return v1.Container{
		Name:            c.Name(),
		ImagePullPolicy: v1.PullIfNotPresent,
		Image:           agnhostImage,
		Command:         cmd,
		SecurityContext: &v1.SecurityContext{},
		Ports: []v1.ContainerPort{
			{
				ContainerPort: int32(c.Port),
				Name:          c.ContainerPortName(),
				Protocol:      c.Protocol,
			},
		},
	}
}
