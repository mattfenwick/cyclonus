package types

import (
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

type Probe struct {
	Ingress  *Table
	Egress   *Table
	Combined *Table
}

type Collector interface {
	RunProbe(request *Request) *Probe
}

type SimulatedCollector struct {
	Policies *matcher.Policy
}

func (s *SimulatedCollector) RunProbe(request *Request) *Probe {
	resources := request.Resources

	ingress := resources.NewTable()
	egress := resources.NewTable()
	combined := resources.NewTable()

	logrus.Infof("running synthetic probe on port %s, protocol %s", request.Port.String(), request.Protocol)

	for _, podFrom := range resources.Pods {
		for _, podTo := range resources.Pods {
			traffic := &matcher.Traffic{
				Source: &matcher.TrafficPeer{
					Internal: &matcher.InternalPeer{
						PodLabels:       podFrom.Labels,
						NamespaceLabels: resources.Namespaces[podFrom.Namespace],
						Namespace:       podFrom.Namespace,
					},
					IP: podFrom.IP,
				},
				Destination: &matcher.TrafficPeer{
					Internal: &matcher.InternalPeer{
						PodLabels:       podTo.Labels,
						NamespaceLabels: resources.Namespaces[podTo.Namespace],
						Namespace:       podTo.Namespace,
					},
					IP: podTo.IP,
				},
				PortProtocol: &matcher.PortProtocol{
					Protocol: request.Protocol,
					Port:     request.Port,
				},
			}

			fr := podFrom.PodString().String()
			to := podTo.PodString().String()
			allowed := s.Policies.IsTrafficAllowed(traffic)
			// TODO could also keep the whole `allowed` struct somewhere

			if allowed.Egress.IsAllowed() {
				egress.Set(fr, to, ConnectivityAllowed)
			} else {
				egress.Set(fr, to, ConnectivityBlocked)
			}

			var portInt int
			var err error
			switch request.Port.Type {
			case intstr.Int:
				portInt = int(request.Port.IntVal)
			case intstr.String:
				portInt, err = podTo.ResolveNamedPort(request.Port.StrVal)
				if err != nil {
					ingress.Set(fr, to, ConnectivityInvalidNamedPort)
					combined.Set(fr, to, ConnectivityInvalidNamedPort)
					continue
				}
			}

			if !podTo.IsServingPortProtocol(portInt, request.Protocol) {
				ingress.Set(fr, to, ConnectivityInvalidPortProtocol)
				combined.Set(fr, to, ConnectivityInvalidPortProtocol)
				continue
			}

			if !allowed.Ingress.IsAllowed() {
				ingress.Set(fr, to, ConnectivityBlocked)
				combined.Set(fr, to, ConnectivityBlocked)
			} else {
				ingress.Set(fr, to, ConnectivityAllowed)
				if allowed.Egress.IsAllowed() {
					combined.Set(fr, to, ConnectivityAllowed)
				} else {
					combined.Set(fr, to, ConnectivityBlocked)
				}
			}
		}
	}

	return &Probe{
		Ingress:  ingress,
		Egress:   egress,
		Combined: combined,
	}
}

type KubeCollector struct {
	Kubernetes      *kube.Kubernetes
	NumberOfWorkers int
}

func (k *KubeCollector) RunProbe(request *Request) *Table {
	podCount := len(request.Resources.Pods)
	size := podCount * podCount

	jobsChan := make(chan *Job, size)
	resultsChan := make(chan *JobResult, size)
	for i := 0; i < k.NumberOfWorkers; i++ {
		go probeWorker(k.Kubernetes, jobsChan, resultsChan)
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

	table := request.Resources.NewTable()
	for _, result := range resultsSlice {
		job := result.Job
		table.Set(job.FromPod.PodString().String(), job.ToPod.PodString().String(), result.Connectivity)
	}
	return table
}

// probeWorker continues polling a pod connectivity status, until the incoming "jobs" channel is closed, and writes results back out to the "results" channel.
// it only writes pass/fail status to a channel and has no failure side effects, this is by design since we do not want to fail inside a goroutine.
func probeWorker(k8s *kube.Kubernetes, jobs <-chan *Job, results chan<- *JobResult) {
	for job := range jobs {
		var result *JobResult
		if job.InvalidNamedPort {
			result = &JobResult{
				Job:          job,
				Connectivity: ConnectivityInvalidNamedPort,
				Err:          errors.Errorf("invalid named port"),
			}
		} else if job.InvalidPortProtocol {
			result = &JobResult{
				Job:          job,
				Connectivity: ConnectivityInvalidPortProtocol,
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

func probeConnectivity(k8s *kube.Kubernetes, job *Job) (Connectivity, string, error) {
	commandDebugString := strings.Join(job.KubeExecCommand(), " ")
	stdout, stderr, commandErr, err := k8s.ExecuteRemoteCommand(job.FromPod.Namespace, job.FromPod.Name, job.FromContainer(), job.ClientCommand())
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
