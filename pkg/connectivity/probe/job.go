package probe

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	v1 "k8s.io/api/core/v1"
	"net"
	"strconv"
	"strings"
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

	TimeoutSeconds int
}

func (j *Job) Key() string {
	return fmt.Sprintf("%s/%s/%s/%s/%s/%d", j.FromKey, j.FromContainer, j.ToKey, j.ToContainer, j.Protocol, j.ResolvedPort)
}

func (j *Job) ToAddress() string {
	return net.JoinHostPort(j.ToHost, strconv.Itoa(j.ResolvedPort))
}

func (j *Job) ClientCommand() []string {
	return []string{"/agnhost", "connect", j.ToAddress(),
		fmt.Sprintf("--timeout=%ds", j.TimeoutSeconds),
		fmt.Sprintf("--protocol=%s", strings.ToLower(string(j.Protocol)))}
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
