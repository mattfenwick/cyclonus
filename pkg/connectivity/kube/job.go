package kube

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

type JobResults struct {
	Job         *Job
	IsConnected bool
	Err         error
	Command     string
}

type Job struct {
	FromPod *Pod
	ToPod   *Pod
}

func (j *Job) ToAddress() string {
	return fmt.Sprintf("%s:%d", kube.QualifiedServiceAddress(j.ToPod.ServiceName(), j.ToPod.Namespace), j.ToPod.Port)
}

func (j *Job) ClientCommand() []string {
	switch j.ToPod.Protocol {
	case v1.ProtocolSCTP:
		return []string{"/agnhost", "connect", j.ToAddress(), "--timeout=1s", "--protocol=sctp"}
	case v1.ProtocolTCP:
		return []string{"/agnhost", "connect", j.ToAddress(), "--timeout=1s", "--protocol=tcp"}
	case v1.ProtocolUDP:
		return []string{"nc", "-v", "-z", "-w", "1", "-u", kube.QualifiedServiceAddress(j.ToPod.ServiceName(), j.ToPod.Namespace), fmt.Sprintf("%d", j.ToPod.Port)}
	default:
		panic(errors.Errorf("protocol %s not supported", j.ToPod.Protocol))
	}
}

func (j *Job) KubeExecCommand() []string {
	return append([]string{
		"kubectl", "exec",
		j.FromPod.Name,
		"-c", j.FromPod.ContainerName,
		"-n", j.FromPod.Namespace,
		"--",
	},
		j.ClientCommand()...)
}

func (j *Job) ToURL() string {
	return fmt.Sprintf("http://%s:%d", j.ToAddress(), j.ToPod.Port)
}
