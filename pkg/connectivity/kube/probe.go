package kube

import (
	"github.com/mattfenwick/cyclonus/pkg/connectivity/types"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strings"
)

func RunKubeProbe(k8s *kube.Kubernetes, request *Request) *Results {
	podCount := len(request.Resources.Pods)
	size := podCount * podCount

	jobsChan := make(chan *Job, size)
	resultsChan := make(chan *JobResult, size)
	for i := 0; i < request.NumberOfWorkers; i++ {
		go probeWorker(k8s, jobsChan, resultsChan)
	}
	jobs := request.Resources.GetJobsForSpecificPortProtocol(request.Port, request.Protocol)
	if len(jobs.BadNamedPort) > 0 || len(jobs.BadPortProtocol) > 0 {
		panic(errors.Errorf("TODO -- handle this better.  Unable to resolve named port OR unable to find port/protocol on pod"))
	}
	for _, job := range jobs.Valid {
		jobsChan <- job
	}
	close(jobsChan)

	var resultsSlice []*JobResult
	for i := 0; i < size; i++ {
		result := <-resultsChan
		resultsSlice = append(resultsSlice, result)
	}
	return &Results{
		Request: request,
		Results: resultsSlice,
	}
}

// probeWorker continues polling a pod connectivity status, until the incoming "jobs" channel is closed, and writes results back out to the "results" channel.
// it only writes pass/fail status to a channel and has no failure side effects, this is by design since we do not want to fail inside a goroutine.
func probeWorker(k8s *kube.Kubernetes, jobs <-chan *Job, results chan<- *JobResult) {
	for job := range jobs {
		var result *JobResult
		if job.InvalidNamedPort {
			result = &JobResult{
				Job:          job,
				Connectivity: types.ConnectivityInvalidNamedPort,
				Err:          errors.Errorf("invalid named port"),
			}
		} else if job.InvalidPortProtocol {
			result = &JobResult{
				Job:          job,
				Connectivity: types.ConnectivityInvalidPortProtocol,
				Err:          errors.Errorf("invalid numbered port or protocol"),
			}
		} else {
			connected, command, err := probeConnectivity(k8s, job)
			result = &JobResult{
				Job:          job,
				Connectivity: connected,
				Err:          err,
				Command:      command,
			}
		}
		results <- result
	}
}

func probeConnectivity(k8s *kube.Kubernetes, job *Job) (types.Connectivity, string, error) {
	commandDebugString := strings.Join(job.KubeExecCommand(), " ")
	stdout, stderr, commandErr, err := k8s.ExecuteRemoteCommand(job.FromPod.Namespace, job.FromPod.Name, job.FromContainer(), job.ClientCommand())
	log.Debugf("stdout, stderr from %s: \n%s\n%s", commandDebugString, stdout, stderr)
	if err != nil {
		log.Errorf("unable to set up command %s: %+v", commandDebugString, err)
		return types.ConnectivityCheckFailed, commandDebugString, nil
	}
	if commandErr != nil {
		log.Debugf("unable to run command %s: %+v", commandDebugString, commandErr)
		return types.ConnectivityBlocked, commandDebugString, nil
	}
	return types.ConnectivityAllowed, commandDebugString, nil
}
