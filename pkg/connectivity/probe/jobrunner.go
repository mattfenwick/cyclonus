package probe

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

type Probe struct {
	Ingress  *Table
	Egress   *Table
	Combined *Table
}

type ProbeRunner struct {
	JobRunner ProbeJobRunner
	Workers   int
}

func NewSimulatedProbeRunner(policies *matcher.Policy) *ProbeRunner {
	return &ProbeRunner{JobRunner: &SimulatedProbeJobRunner{Policies: policies}, Workers: 1}
}

func NewKubeProbeRunner(kubernetes *kube.Kubernetes, workers int) *ProbeRunner {
	return &ProbeRunner{JobRunner: &KubeProbeJobRunner{Kubernetes: kubernetes}, Workers: workers}
}

func (p *ProbeRunner) RunProbeForConfig(probeConfig *generator.ProbeConfig, resources *Resources) *Probe {
	if probeConfig.AllAvailable {
		return p.RunAllAvailablePortsProbe(resources)
	} else if probeConfig.PortProtocol != nil {
		return p.RunProbeFixedPortProtocol(resources, probeConfig.PortProtocol.Port, probeConfig.PortProtocol.Protocol)
	} else {
		panic(errors.Errorf("invalid ProbeConfig value %+v", probeConfig))
	}
}

func (p *ProbeRunner) RunAllAvailablePortsProbe(resources *Resources) *Probe {
	probe := &Probe{Ingress: resources.NewTable(), Egress: resources.NewTable(), Combined: resources.NewTable()}
	for _, result := range p.runProbe(resources.GetJobsAllAvailableServers()) {
		fr := result.Job.FromKey
		to := result.Job.ToKey
		key := fmt.Sprintf("%s/%d", result.Job.Protocol, result.Job.ResolvedPort)
		if result.Ingress != nil {
			probe.Ingress.Set(fr, to, key, *result.Ingress)
		}
		if result.Egress != nil {
			probe.Egress.Set(fr, to, key, *result.Egress)
		}
		probe.Combined.Set(fr, to, key, result.Combined)
	}
	return probe
}

func (p *ProbeRunner) RunProbeFixedPortProtocol(resources *Resources, port intstr.IntOrString, protocol v1.Protocol) *Probe {
	jobs := resources.GetJobsForNamedPortProtocol(port, protocol)
	probe := &Probe{Ingress: resources.NewTable(), Egress: resources.NewTable(), Combined: resources.NewTable()}
	for _, result := range p.runProbe(jobs) {
		fr := result.Job.FromKey
		to := result.Job.ToKey
		if result.Ingress != nil {
			probe.Ingress.Set(fr, to, "", *result.Ingress)
		}
		if result.Egress != nil {
			probe.Egress.Set(fr, to, "", *result.Egress)
		}
		probe.Combined.Set(fr, to, "", result.Combined)
	}
	return probe
}

func (p *ProbeRunner) runProbe(jobs *Jobs) []*JobResult {
	size := len(jobs.Valid)
	jobsChan := make(chan *Job, size)
	resultsChan := make(chan *JobResult, size)
	for i := 0; i < p.Workers; i++ {
		go p.worker(jobsChan, resultsChan)
	}
	for _, job := range jobs.Valid {
		jobsChan <- job
	}
	close(jobsChan)

	var resultSlice []*JobResult
	for i := 0; i < size; i++ {
		result := <-resultsChan
		resultSlice = append(resultSlice, result)
	}
	invalidPP := ConnectivityInvalidPortProtocol
	for _, j := range jobs.BadPortProtocol {
		resultSlice = append(resultSlice, &JobResult{
			Job:      j,
			Ingress:  &invalidPP,
			Combined: ConnectivityInvalidPortProtocol,
		})
	}

	invalidNamedPort := ConnectivityInvalidNamedPort
	for _, j := range jobs.BadNamedPort {
		resultSlice = append(resultSlice, &JobResult{
			Job:      j,
			Ingress:  &invalidNamedPort,
			Combined: ConnectivityInvalidNamedPort,
		})
	}

	return resultSlice
}

// probeWorker continues polling a pod connectivity status, until the incoming "jobs" channel is closed, and writes results back out to the "results" channel.
// it only writes pass/fail status to a channel and has no failure side effects, this is by design since we do not want to fail inside a goroutine.
func (p *ProbeRunner) worker(jobs <-chan *Job, results chan<- *JobResult) {
	for job := range jobs {
		results <- p.JobRunner.RunJob(job)
	}
}

type ProbeJobRunner interface {
	RunJob(job *Job) *JobResult
}

type SimulatedProbeJobRunner struct {
	Policies *matcher.Policy
}

func (s *SimulatedProbeJobRunner) RunJob(job *Job) *JobResult {
	allowed := s.Policies.IsTrafficAllowed(job.Traffic())
	// TODO could also keep the whole `allowed` struct somewhere

	logrus.Tracef("to %s\n%s\n", utils.JsonString(job), allowed.Table())

	var combined, ingress, egress = ConnectivityBlocked, ConnectivityBlocked, ConnectivityBlocked
	if allowed.Ingress.IsAllowed() {
		ingress = ConnectivityAllowed
	}
	if allowed.Egress.IsAllowed() {
		egress = ConnectivityAllowed
	}
	if allowed.IsAllowed() {
		combined = ConnectivityAllowed
	}

	return &JobResult{Job: job, Ingress: &ingress, Egress: &egress, Combined: combined}
}

type KubeProbeJobRunner struct {
	Kubernetes *kube.Kubernetes
}

func (k *KubeProbeJobRunner) RunJob(job *Job) *JobResult {
	connectivity, _, _ := probeConnectivity(k.Kubernetes, job)
	return &JobResult{Job: job, Ingress: nil, Egress: nil, Combined: connectivity}
}

func probeConnectivity(k8s *kube.Kubernetes, job *Job) (Connectivity, string, error) {
	commandDebugString := strings.Join(job.KubeExecCommand(), " ")
	stdout, stderr, commandErr, err := k8s.ExecuteRemoteCommand(job.FromNamespace, job.FromPod, job.FromContainer, job.ClientCommand())
	logrus.Debugf("stdout, stderr from %s: \n%s\n%s", commandDebugString, stdout, stderr)
	if err != nil {
		logrus.Errorf("unable to set up command %s: %+v", commandDebugString, err)
		return ConnectivityCheckFailed, commandDebugString, nil
	}
	if commandErr != nil {
		logrus.Debugf("unable to run command %s: %+v", commandDebugString, commandErr)
		return ConnectivityBlocked, commandDebugString, nil
	}
	return ConnectivityAllowed, commandDebugString, nil
}
