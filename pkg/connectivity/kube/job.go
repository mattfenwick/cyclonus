package kube

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/types"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

type JobResult struct {
	Job          *Job
	Connectivity types.Connectivity
	Err          error
	Command      string
}

type Job struct {
	FromPod             *Pod
	ToPod               *Pod
	Port                int
	Protocol            v1.Protocol
	InvalidNamedPort    bool
	InvalidPortProtocol bool
}

func (j *Job) FromContainer() string {
	return j.FromPod.KubePod.Spec.Containers[0].Name
}

func (j *Job) ToAddress() string {
	return fmt.Sprintf("%s:%d", kube.QualifiedServiceAddress(j.ToPod.ServiceName(), j.ToPod.Namespace), j.Port)
}

func (j *Job) ClientCommand() []string {
	switch j.Protocol {
	case v1.ProtocolSCTP:
		return []string{"/agnhost", "connect", j.ToAddress(), "--timeout=1s", "--protocol=sctp"}
	case v1.ProtocolTCP:
		return []string{"/agnhost", "connect", j.ToAddress(), "--timeout=1s", "--protocol=tcp"}
	case v1.ProtocolUDP:
		return []string{"nc", "-v", "-z", "-w", "1", "-u", kube.QualifiedServiceAddress(j.ToPod.ServiceName(), j.ToPod.Namespace), fmt.Sprintf("%d", j.Port)}
	default:
		panic(errors.Errorf("protocol %s not supported", j.Protocol))
	}
}

func (j *Job) KubeExecCommand() []string {
	return append([]string{
		"kubectl", "exec",
		j.FromPod.Name,
		"-c", j.FromContainer(),
		"-n", j.FromPod.Namespace,
		"--",
	},
		j.ClientCommand()...)
}

func (j *Job) ToURL() string {
	return fmt.Sprintf("http://%s:%d", j.ToAddress(), j.Port)
}
