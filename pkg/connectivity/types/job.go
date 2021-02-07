package types

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Jobs struct {
	Valid           []*Job
	BadNamedPort    []*Job
	BadPortProtocol []*Job
}

type JobResult struct {
	Job      *Job
	Ingress  *Connectivity
	Egress   *Connectivity
	Combined Connectivity
}

//func (j *JobResult) Connectivity() Connectivity {
//	return CombineIngressEgressConnectivity(j.Ingress, j.Egress)
//}

type Job struct {
	FromKey             string
	FromNamespace       string
	FromNamespaceLabels map[string]string
	FromPod             string
	FromPodLabels       map[string]string
	FromContainer       string
	FromIP              string

	ToKey             string
	ToHost            string
	ToNamespace       string
	ToNamespaceLabels map[string]string
	ToPodLabels       map[string]string
	ToIP              string

	Port     int
	Protocol v1.Protocol
}

func (j *Job) ToAddress() string {
	return fmt.Sprintf("%s:%d", j.ToHost, j.Port)
}

func (j *Job) ClientCommand() []string {
	switch j.Protocol {
	case v1.ProtocolSCTP:
		return []string{"/agnhost", "connect", j.ToAddress(), "--timeout=1s", "--protocol=sctp"}
	case v1.ProtocolTCP:
		return []string{"/agnhost", "connect", j.ToAddress(), "--timeout=1s", "--protocol=tcp"}
	case v1.ProtocolUDP:
		return []string{"nc", "-v", "-z", "-w", "1", "-u", j.ToHost, fmt.Sprintf("%d", j.Port)}
	default:
		panic(errors.Errorf("protocol %s not supported", j.Protocol))
	}
}

func (j *Job) KubeExecCommand() []string {
	return append([]string{
		"kubectl", "exec",
		j.FromPod,
		"-c", j.FromContainer,
		"-n", j.FromNamespace,
		"--",
	},
		j.ClientCommand()...)
}

// TODO why is this here?
//func (j *Job) ToURL() string {
//	return fmt.Sprintf("http://%s:%d", j.ToAddress(), j.Port)
//}

func (j *Job) Traffic() *matcher.Traffic {
	return &matcher.Traffic{
		Source: &matcher.TrafficPeer{
			Internal: &matcher.InternalPeer{
				PodLabels:       j.FromPodLabels,
				NamespaceLabels: j.FromNamespaceLabels,
				Namespace:       j.FromNamespace,
			},
			IP: j.FromIP,
		},
		Destination: &matcher.TrafficPeer{
			Internal: &matcher.InternalPeer{
				PodLabels:       j.ToPodLabels,
				NamespaceLabels: j.ToNamespaceLabels,
				Namespace:       j.ToNamespace,
			},
			IP: j.ToIP,
		},
		PortProtocol: &matcher.PortProtocol{
			Protocol: j.Protocol,
			Port:     intstr.FromInt(j.Port),
		},
	}
}
