package types

import (
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/sirupsen/logrus"
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

func (p *ProbeRunner) RunProbe(jobs *Jobs, newTable func() *Table) *Probe {
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

	probe := &Probe{Ingress: newTable(), Egress: newTable(), Combined: newTable()}
	buildTable(probe, jobs, resultSlice)
	return probe
}

func buildTable(probe *Probe, jobs *Jobs, results []*JobResult) {
	for _, j := range jobs.BadPortProtocol {
		probe.Ingress.Set(j.FromKey, j.ToKey, ConnectivityInvalidPortProtocol)
		probe.Combined.Set(j.FromKey, j.ToKey, ConnectivityInvalidPortProtocol)
	}

	for _, j := range jobs.BadNamedPort {
		probe.Ingress.Set(j.FromKey, j.ToKey, ConnectivityInvalidNamedPort)
		probe.Combined.Set(j.FromKey, j.ToKey, ConnectivityInvalidNamedPort)
	}

	for _, result := range results {
		fr := result.Job.FromKey
		to := result.Job.ToKey
		if result.Ingress != nil {
			probe.Ingress.Set(fr, to, *result.Ingress)
		}
		if result.Egress != nil {
			probe.Egress.Set(fr, to, *result.Egress)
		}
		probe.Combined.Set(fr, to, result.Combined)
	}
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
