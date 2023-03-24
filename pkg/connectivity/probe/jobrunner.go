package probe

import (
	"strings"

	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/mattfenwick/cyclonus/pkg/worker"
	"github.com/sirupsen/logrus"
)

type Runner struct {
	JobRunner  JobRunner
	JobBuilder *JobBuilder
}

func NewSimulatedRunner(policies *matcher.Policy, jobBuilder *JobBuilder) *Runner {
	return &Runner{JobRunner: &SimulatedJobRunner{Policies: policies}, JobBuilder: jobBuilder}
}

func NewKubeRunner(kubernetes kube.IKubernetes, workers int, jobBuilder *JobBuilder) *Runner {
	return &Runner{JobRunner: &KubeJobRunner{Kubernetes: kubernetes, Workers: workers}, JobBuilder: jobBuilder}
}

func NewKubeBatchRunner(kubernetes kube.IKubernetes, workers int, jobBuilder *JobBuilder) *Runner {
	return &Runner{JobRunner: NewKubeBatchJobRunner(kubernetes, workers), JobBuilder: jobBuilder}
}

func (p *Runner) RunProbeForConfig(probeConfig *generator.ProbeConfig, resources *Resources) *Table {
	jobs := p.JobBuilder.GetJobsForProbeConfig(resources, probeConfig)
	logrus.Debugf("got jobs %+v", jobs)
	jobresults := p.runProbe(jobs)
	if probeConfig.Mode == generator.ProbeModeNodeIP {
		return NewNodeTableFromJobResults(resources, jobresults)
	} else {
		return NewPodTableFromJobResults(resources, jobresults)
	}
}

func (p *Runner) runProbe(jobs *Jobs) []*JobResult {
	logrus.Debugf("running probe for job %+v", jobs)
	resultSlice := p.JobRunner.RunJobs(jobs.Valid)

	for _, res := range resultSlice {
		logrus.Debugf("resultslice combined: %+v, ingress: %+v, egress %+v", res.Combined, res.Ingress, res.Egress)
	}

	invalidPP := ConnectivityInvalidPortProtocol
	unknown := ConnectivityUnknown
	for _, j := range jobs.BadPortProtocol {
		resultSlice = append(resultSlice, &JobResult{
			Job:      j,
			Ingress:  &invalidPP,
			Egress:   &unknown,
			Combined: ConnectivityInvalidPortProtocol,
		})
	}

	invalidNamedPort := ConnectivityInvalidNamedPort
	for _, j := range jobs.BadNamedPort {
		resultSlice = append(resultSlice, &JobResult{
			Job:      j,
			Ingress:  &invalidNamedPort,
			Egress:   &unknown,
			Combined: ConnectivityInvalidNamedPort,
		})
	}

	return resultSlice
}

type JobRunner interface {
	RunJobs(job []*Job) []*JobResult
}

type SimulatedJobRunner struct {
	Policies *matcher.Policy
}

func (s *SimulatedJobRunner) RunJobs(jobs []*Job) []*JobResult {
	results := make([]*JobResult, len(jobs))
	for i, job := range jobs {
		results[i] = s.RunJob(job)
	}
	return results
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
	Kubernetes kube.IKubernetes
	Workers    int
}

func (k *KubeJobRunner) RunJobs(jobs []*Job) []*JobResult {
	logrus.Debugf("run job single with %+v", jobs)
	size := len(jobs)
	jobsChan := make(chan *Job, size)
	resultsChan := make(chan *JobResult, size)
	for i := 0; i < k.Workers; i++ {
		go k.worker(jobsChan, resultsChan)
	}
	for _, job := range jobs {
		jobsChan <- job
	}
	close(jobsChan)

	var resultSlice []*JobResult
	for i := 0; i < size; i++ {
		result := <-resultsChan
		resultSlice = append(resultSlice, result)
	}

	return resultSlice
}

// probeWorker continues polling a pod connectivity status, until the incoming "jobs" channel is closed, and writes results back out to the "results" channel.
// it only writes pass/fail status to a channel and has no failure side effects, this is by design since we do not want to fail inside a goroutine.
func (k *KubeJobRunner) worker(jobs <-chan *Job, results chan<- *JobResult) {
	for job := range jobs {
		logrus.Debugf("probing connectivity for job %+v", job)
		connectivity, _ := probeConnectivity(k.Kubernetes, job)
		results <- &JobResult{
			Job:      job,
			Combined: connectivity,
		}
	}
}

func probeConnectivity(k8s kube.IKubernetes, job *Job) (Connectivity, string) {
	commandDebugString := strings.Join(job.KubeExecCommand(), " ")
	stdout, stderr, commandErr, err := k8s.ExecuteRemoteCommand(job.FromNamespace, job.FromPod, job.FromContainer, job.ClientCommand())
	logrus.Debugf("stdout, stderr from [%s]: \n%s\n%s", commandDebugString, stdout, stderr)
	if err != nil {
		logrus.Errorf("unable to set up command [%s]: %+v", commandDebugString, err)
		return ConnectivityCheckFailed, commandDebugString
	}
	if commandErr != nil {
		logrus.Debugf("unable to run command [%s]: %+v", commandDebugString, commandErr)
		return ConnectivityBlocked, commandDebugString
	}
	return ConnectivityAllowed, commandDebugString
}

type KubeBatchJobRunner struct {
	Client  *worker.Client
	Workers int
}

func NewKubeBatchJobRunner(k8s kube.IKubernetes, workers int) *KubeBatchJobRunner {
	return &KubeBatchJobRunner{Client: &worker.Client{Kubernetes: k8s}, Workers: workers}
}

func (k *KubeBatchJobRunner) RunJobs(jobs []*Job) []*JobResult {
	jobMap := map[string]*Job{}

	// 1. batch up jobs
	batches := map[string]*worker.Batch{}
	for _, job := range jobs {
		ns, pod := job.FromNamespace, job.FromPod
		if _, ok := batches[job.FromKey]; !ok {
			batches[job.FromKey] = &worker.Batch{Namespace: ns, Pod: pod, Container: job.FromContainer}
		}
		batch := batches[job.FromKey]
		batch.Requests = append(batch.Requests, &worker.Request{
			Key:      job.Key(),
			Protocol: job.Protocol,
			Host:     job.ToHost,
			Port:     job.ResolvedPort,
		})

		jobMap[job.Key()] = job
	}

	// 2. send them out and get the results
	size := len(jobs)
	batchChan := make(chan *worker.Batch, size)
	resultsChan := make(chan *JobResult, size)
	for i := 0; i < k.Workers; i++ {
		go k.worker(jobMap, batchChan, resultsChan)
	}
	for _, b := range batches {
		batchChan <- b
	}
	close(batchChan)

	var jobResults []*JobResult
	for i := 0; i < size; i++ {
		result := <-resultsChan
		jobResults = append(jobResults, result)
	}

	return jobResults
}

func (k *KubeBatchJobRunner) worker(jobMap map[string]*Job, batches <-chan *worker.Batch, jobResults chan<- *JobResult) {
	for b := range batches {
		results, err := k.Client.Batch(b)
		if err != nil {
			logrus.Errorf("unable to issue batch request: %+v", err)
			for _, r := range b.Requests {
				jobResults <- &JobResult{
					Job:      jobMap[r.Key],
					Combined: ConnectivityCheckFailed,
				}
			}
		} else {
			for _, r := range results {
				var c Connectivity
				if r.IsSuccess() {
					c = ConnectivityAllowed
				} else {
					logrus.Debugf("request to %s failed: %s", r.Request.Key, r.Error)
					c = ConnectivityBlocked
				}
				jobResults <- &JobResult{
					Job:      jobMap[r.Request.Key],
					Combined: c,
				}
			}
		}
	}
}
