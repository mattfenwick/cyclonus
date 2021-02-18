package probe

import (
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/mattfenwick/cyclonus/pkg/worker"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

type Runner struct {
	JobRunner JobRunner
	Workers   int
}

func NewSimulatedRunner(policies *matcher.Policy) *Runner {
	return &Runner{JobRunner: &SimulatedJobRunner{Policies: policies}, Workers: 1}
}

func NewKubeRunner(kubernetes *kube.Kubernetes, workers int) *Runner {
	return &Runner{JobRunner: &KubeJobRunner{Kubernetes: kubernetes}, Workers: workers}
}

func (p *Runner) RunProbeForConfig(probeConfig *generator.ProbeConfig, resources *Resources) *Table {
	if probeConfig.AllAvailable {
		return p.RunAllAvailablePortsProbe(resources)
	} else if probeConfig.PortProtocol != nil {
		return p.RunProbeFixedPortProtocol(resources, probeConfig.PortProtocol.Port, probeConfig.PortProtocol.Protocol)
	} else {
		panic(errors.Errorf("invalid ProbeConfig value %+v", probeConfig))
	}
}

func (p *Runner) RunAllAvailablePortsProbe(resources *Resources) *Table {
	return NewTableFromJobResults(resources, p.runProbe(resources.GetJobsAllAvailableServers()))
}

func (p *Runner) RunProbeFixedPortProtocol(resources *Resources, port intstr.IntOrString, protocol v1.Protocol) *Table {
	return NewTableFromJobResults(resources, p.runProbe(resources.GetJobsForNamedPortProtocol(port, protocol)))
}

func (p *Runner) runProbe(jobs *Jobs) []*JobResult {
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
func (p *Runner) worker(jobs <-chan *Job, results chan<- *JobResult) {
	for job := range jobs {
		results <- p.JobRunner.RunJob(job)
	}
}

type JobRunner interface {
	RunJob(job *Job) *JobResult
}

type SimulatedJobRunner struct {
	Policies *matcher.Policy
}

func (s *SimulatedJobRunner) RunJob(job *Job) *JobResult {
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

type KubeJobRunner struct {
	Kubernetes *kube.Kubernetes
}

func (k *KubeJobRunner) RunJob(job *Job) *JobResult {
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

type KubeBatchJobRunner struct {
	Client *worker.Client
}

func NewKubeBatchJobRunner(k8s *kube.Kubernetes) *KubeBatchJobRunner {
	return &KubeBatchJobRunner{
		Client: &worker.Client{Kubernetes: k8s},
	}
}

func (k *KubeBatchJobRunner) RunJobs(jobs []*Job) []*JobResult {
	jobMap := map[string]*Job{}

	// 1. batch up jobs
	batches := map[string]map[string]*worker.Batch{}
	for _, job := range jobs {
		ns, pod := job.FromNamespace, job.FromPod
		if _, ok := batches[ns]; !ok {
			batches[ns] = map[string]*worker.Batch{}
		}
		if _, ok := batches[ns][pod]; !ok {
			batches[ns][pod] = &worker.Batch{Namespace: ns, Pod: pod, Container: job.FromContainer}
		}
		batch := batches[ns][pod]
		batch.Requests = append(batch.Requests, &worker.Request{
			Key:      job.Key(),
			Protocol: job.Protocol,
			Host:     job.ToHost,
			Port:     job.ResolvedPort,
		})

		jobMap[job.Key()] = job
	}

	var jobResults []*JobResult
	// 2. send them out and get the results
	for _, dict := range batches {
		for _, b := range dict {
			results, err := k.Client.Batch(b)
			if err != nil {
				for _, r := range b.Requests {
					jobResults = append(jobResults, &JobResult{
						Job:      jobMap[r.Key],
						Combined: ConnectivityCheckFailed,
					})
				}
			} else {
				for _, r := range results {
					var c Connectivity
					if r.IsSuccess() {
						c = ConnectivityAllowed
					} else {
						c = ConnectivityBlocked
					}
					jobResults = append(jobResults, &JobResult{
						Job:      jobMap[r.Request.Key],
						Combined: c,
					})
				}
			}
		}
	}

	return jobResults
}
