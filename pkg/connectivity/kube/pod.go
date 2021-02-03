package kube

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

const (
	agnhostImage = "k8s.gcr.io/e2e-test-images/agnhost:2.21"
)

type Pod struct {
	Namespace string
	Name      string
	Labels    map[string]string
	Ports     []int
	Protocols []v1.Protocol
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
		Ports:     ports,
		Protocols: protocols,
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

	for _, port := range p.Ports {
		for _, protocol := range p.Protocols {
			service.Spec.Ports = append(service.Spec.Ports, v1.ServicePort{
				Name:     fmt.Sprintf("service-port-%s-%d", strings.ToLower(string(protocol)), port),
				Protocol: protocol,
				Port:     int32(port),
			})
		}
	}

	return service
}

func (p *Pod) KubeContainers() []v1.Container {
	var containers []v1.Container
	for _, port := range p.Ports {
		for _, protocol := range p.Protocols {
			var cmd []string

			switch protocol {
			case v1.ProtocolTCP:
				cmd = []string{"/agnhost", "serve-hostname", "--tcp", "--http=false", "--port", fmt.Sprintf("%d", port)}
			case v1.ProtocolUDP:
				cmd = []string{"/agnhost", "serve-hostname", "--udp", "--http=false", "--port", fmt.Sprintf("%d", port)}
			case v1.ProtocolSCTP:
				cmd = []string{"/agnhost", "netexec", "--sctp-port", fmt.Sprintf("%d", port)}
			default:
				panic(errors.Errorf("invalid protocol %s", protocol))
			}
			containers = append(containers, v1.Container{
				Name:            fmt.Sprintf("cont-%d-%s", port, strings.ToLower(string(protocol))),
				ImagePullPolicy: v1.PullIfNotPresent,
				Image:           agnhostImage,
				Command:         cmd,
				SecurityContext: &v1.SecurityContext{},
				Ports: []v1.ContainerPort{
					{
						ContainerPort: int32(port),
						Name:          fmt.Sprintf("serve-%d-%s", port, strings.ToLower(string(protocol))),
						Protocol:      protocol,
					},
				},
			})
		}
	}
	return containers
}

func (p *Pod) ResolvePort(port intstr.IntOrString) (int, error) {
	switch port.Type {
	case intstr.Int:
		return int(port.IntVal), nil
	case intstr.String:
		for _, c := range p.KubePod.Spec.Containers {
			if len(c.Ports) != 1 {
				return 0, errors.Errorf("expected container %s/%s/%s to have 1 port, found %d", p.Namespace, p.Name, c.Name, len(c.Ports))
			}
			if c.Ports[0].Name == port.StrVal {
				return int(c.Ports[0].ContainerPort), nil
			}
		}
		return 0, errors.Errorf("unable to resolve named port %s on pod %s/%s", port.StrVal, p.Namespace, p.Name)
	default:
		return 0, errors.Errorf("invalid intstr.IntOrString value %+v", port)
	}
}
