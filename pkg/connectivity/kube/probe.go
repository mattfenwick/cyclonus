package kube

import (
	"github.com/mattfenwick/cyclonus/pkg/kube"
	log "github.com/sirupsen/logrus"
	"strings"
)

func RunKubeProbe(k8s *kube.Kubernetes, request *Request) *Results {
	podCount := len(request.Model.Pods)
	size := podCount * podCount

	jobs := make(chan *Job, size)
	results := make(chan *JobResults, size)
	for i := 0; i < request.NumberOfWorkers; i++ {
		go probeWorker(k8s, jobs, results)
	}
	for _, job := range request.Model.Jobs {
		jobs <- job
	}
	close(jobs)

	var resultsSlice []*JobResults
	for i := 0; i < size; i++ {
		result := <-results
		resultsSlice = append(resultsSlice, result)
	}
	return &Results{
		Request: request,
		Results: resultsSlice,
	}
}

// probeWorker continues polling a pod connectivity status, until the incoming "jobs" channel is closed, and writes results back out to the "results" channel.
// it only writes pass/fail status to a channel and has no failure side effects, this is by design since we do not want to fail inside a goroutine.
func probeWorker(k8s *kube.Kubernetes, jobs <-chan *Job, results chan<- *JobResults) {
	for job := range jobs {
		connected, command, err := probeConnectivity(k8s, job)
		result := &JobResults{
			Job:         job,
			IsConnected: connected,
			Err:         err,
			Command:     command,
		}
		results <- result
	}
}

func probeConnectivity(k8s *kube.Kubernetes, job *Job) (bool, string, error) {
	commandDebugString := strings.Join(job.KubeExecCommand(), " ")
	stdout, stderr, commandErr, err := k8s.ExecuteRemoteCommand(job.FromPod.Namespace, job.FromPod.Name, job.FromPod.ContainerName, job.ClientCommand())
	log.Debugf("stdout, stderr from %s: \n%s\n%s", commandDebugString, stdout, stderr)
	if err != nil {
		log.Errorf("unable to set up command %s: %+v", commandDebugString, err)
		return false, commandDebugString, nil
	}
	if commandErr != nil {
		log.Debugf("unable to run command %s: %+v", commandDebugString, commandErr)
		return false, commandDebugString, nil
	}
	return true, commandDebugString, nil
}
