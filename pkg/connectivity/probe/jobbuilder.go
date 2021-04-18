package probe

import (
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type JobBuilder struct {
	TimeoutSeconds int
}

func (j *JobBuilder) GetJobsForProbeConfig(resources *Resources, config *generator.ProbeConfig) *Jobs {
	if config.AllAvailable {
		return j.GetJobsAllAvailableServers(resources, config.Mode)
	} else if config.PortProtocol != nil {
		return j.GetJobsForNamedPortProtocol(resources, config.PortProtocol.Port, config.PortProtocol.Protocol, config.Mode)
	} else {
		panic(errors.Errorf("invalid ProbeConfig %+v", config))
	}
}

func (j *JobBuilder) GetJobsForNamedPortProtocol(resources *Resources, port intstr.IntOrString, protocol v1.Protocol, mode generator.ProbeMode) *Jobs {
	jobs := &Jobs{}
	for _, podFrom := range resources.Pods {
		for _, podTo := range resources.Pods {
			job := &Job{
				FromKey:             podFrom.PodString().String(),
				FromNamespace:       podFrom.Namespace,
				FromNamespaceLabels: resources.Namespaces[podFrom.Namespace],
				FromPod:             podFrom.Name,
				FromPodLabels:       podFrom.Labels,
				FromContainer:       podFrom.Containers[0].Name,
				FromIP:              podFrom.IP,
				ToKey:               podTo.PodString().String(),
				ToHost:              podTo.Host(mode),
				ToNamespace:         podTo.Namespace,
				ToNamespaceLabels:   resources.Namespaces[podTo.Namespace],
				ToPodLabels:         podTo.Labels,
				ToIP:                podTo.IP,
				ResolvedPort:        -1,
				ResolvedPortName:    "",
				Protocol:            protocol,
				TimeoutSeconds:      j.TimeoutSeconds,
			}

			switch port.Type {
			case intstr.String:
				job.ResolvedPortName = port.StrVal
				// TODO what about protocol?
				portInt, err := podTo.ResolveNamedPort(port.StrVal)
				if err != nil {
					jobs.BadNamedPort = append(jobs.BadNamedPort, job)
					continue
				}
				job.ResolvedPort = portInt
			case intstr.Int:
				job.ResolvedPort = int(port.IntVal)
				// TODO what about protocol?
				portName, err := podTo.ResolveNumberedPort(int(port.IntVal))
				if err != nil {
					jobs.BadPortProtocol = append(jobs.BadPortProtocol, job)
					continue
				}
				job.ResolvedPortName = portName
			default:
				panic(errors.Errorf("invalid IntOrString value %+v", port))
			}

			jobs.Valid = append(jobs.Valid, job)
		}
	}
	return jobs
}

func (j *JobBuilder) GetJobsAllAvailableServers(resources *Resources, mode generator.ProbeMode) *Jobs {
	var jobs []*Job
	for _, podFrom := range resources.Pods {
		for _, podTo := range resources.Pods {
			for _, contTo := range podTo.Containers {
				jobs = append(jobs, &Job{
					FromKey:             podFrom.PodString().String(),
					FromNamespace:       podFrom.Namespace,
					FromNamespaceLabels: resources.Namespaces[podFrom.Namespace],
					FromPod:             podFrom.Name,
					FromPodLabels:       podFrom.Labels,
					FromContainer:       podFrom.Containers[0].Name,
					FromIP:              podFrom.IP,
					ToKey:               podTo.PodString().String(),
					ToHost:              podTo.Host(mode),
					ToNamespace:         podTo.Namespace,
					ToNamespaceLabels:   resources.Namespaces[podTo.Namespace],
					ToPodLabels:         podTo.Labels,
					ToContainer:         contTo.Name,
					ToIP:                podTo.IP,
					ResolvedPort:        contTo.Port,
					ResolvedPortName:    contTo.PortName,
					Protocol:            contTo.Protocol,
					TimeoutSeconds:      j.TimeoutSeconds,
				})
			}
		}
	}
	return &Jobs{Valid: jobs}
}
