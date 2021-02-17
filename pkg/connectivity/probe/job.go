package probe

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
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

func (jr *JobResult) Key() string {
	return fmt.Sprintf("%s/%d", jr.Job.Protocol, jr.Job.ResolvedPort)
}

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
	ToContainer       string
	ToIP              string

	ResolvedPort     int
	ResolvedPortName string
	Protocol         v1.Protocol
}

func (j *Job) ToAddress() string {
	return fmt.Sprintf("%s:%d", j.ToHost, j.ResolvedPort)
}

func (j *Job) ClientCommand() []string {
	switch j.Protocol {
	case v1.ProtocolSCTP:
		return []string{"/agnhost", "connect", j.ToAddress(), "--timeout=1s", "--protocol=sctp"}
	case v1.ProtocolTCP:
		return []string{"/agnhost", "connect", j.ToAddress(), "--timeout=1s", "--protocol=tcp"}
	case v1.ProtocolUDP:
		return []string{"/agnhost", "connect", j.ToAddress(), "--timeout=1s", "--protocol=udp"}
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
		ResolvedPort:     j.ResolvedPort,
		ResolvedPortName: j.ResolvedPortName,
		Protocol:         j.Protocol,
	}
}
