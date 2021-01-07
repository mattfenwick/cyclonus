package kube

import (
	"bytes"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/netpol"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"strings"
)

type ProbeJob struct {
	FromNamespace  string
	FromPod        string
	FromContainer  string
	ToAddress      string
	ToPort         int
	TimeoutSeconds int
	CommandType    ProbeCommandType
	FromKey        string
	ToKey          string
}

func (pj *ProbeJob) Command() ProbeCommand {
	switch pj.CommandType {
	case ProbeCommandTypeCurl:
		return &CurlCommand{
			TimeoutSeconds: pj.TimeoutSeconds,
			URL:            pj.ToURL(),
		}
	case ProbeCommandTypeWget:
		return &WgetCommand{
			TimeoutSeconds: pj.TimeoutSeconds,
			URL:            pj.ToURL(),
		}
	case ProbeCommandTypeNetcat:
		return &NetcatCommand{
			TimeoutSeconds: pj.TimeoutSeconds,
			ToAddress:      pj.ToAddress,
			ToPort:         pj.ToPort,
		}
	default:
		panic(errors.Errorf("invalid command type '%s'", pj.CommandType))
	}
}

func (pj *ProbeJob) GetFromKey() string {
	if pj.FromKey != "" {
		return pj.FromKey
	}
	return fmt.Sprintf("%s/%s/%s", pj.FromNamespace, pj.FromPod, pj.FromContainer)
}

func (pj *ProbeJob) GetToKey() string {
	if pj.ToKey != "" {
		return pj.ToKey
	}
	return fmt.Sprintf("%s:%d", pj.ToAddress, pj.ToPort)
}

func (pj *ProbeJob) KubeExecCommand() []string {
	return append([]string{
		"kubectl", "exec",
		pj.FromPod,
		"-c", pj.FromContainer,
		"-n", pj.FromNamespace,
		"--",
	},
		pj.Command().Command()...)
}

func (pj *ProbeJob) ToURL() string {
	return fmt.Sprintf("http://%s:%d", pj.ToAddress, pj.ToPort)
}

func (k *Kubernetes) Probe(job *ProbeJob) (*ProbeResult, error) {
	fromNamespace := job.FromNamespace
	fromPod := job.FromPod
	fromContainer := job.FromContainer

	command := job.Command()

	log.Infof("Running: %s", strings.Join(job.KubeExecCommand(), " "))
	out, errorOut, execErr, err := k.ExecuteRemoteCommand(fromNamespace, fromPod, fromContainer, command.Command())
	log.Infof("finished, with out '%s' and errOut '%s'", out, errorOut)

	if err != nil {
		return nil, err
	}
	return command.ParseOutput(out, errorOut, execErr), nil
}

// ExecuteRemoteCommand executes a remote shell command on the given pod
// returns the output from stdout and stderr
func (k *Kubernetes) ExecuteRemoteCommand(namespace string, pod string, container string, command []string) (string, string, error, error) {
	kubeCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	restCfg, err := kubeCfg.ClientConfig()
	if err != nil {
		return "", "", nil, errors.Wrapf(err, "unable to get rest config from kube config")
	}

	request := k.ClientSet.
		CoreV1().
		RESTClient().
		Post().
		Namespace(namespace).
		Resource("pods").
		Name(pod).
		SubResource("exec").
		VersionedParams(
			&v1.PodExecOptions{
				Container: container,
				Command:   command,
				Stdin:     false,
				Stdout:    true,
				Stderr:    true,
				TTY:       true,
			},
			scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(restCfg, "POST", request.URL())
	if err != nil {
		return "", "", nil, errors.Wrapf(err, "unable to instantiate SPDYExecutor")
	}

	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: buf,
		Stderr: errBuf,
	})

	out, errOut := buf.String(), errBuf.String()
	return out, errOut, err, nil
}

// ProbeJobResult : if command can't be run, Result should be nil and Err non-nil
type ProbeJobResult struct {
	Job    *ProbeJob
	Result *ProbeResult
	Err    error
}

func (k *Kubernetes) ProbeConnectivity(jobs []*ProbeJob) *netpol.StringTruthTable {
	log.Infof("running %d probe jobs", len(jobs))

	numberOfWorkers := 30
	jobsChan := make(chan *ProbeJob, len(jobs))
	results := make(chan *ProbeJobResult, len(jobs))
	for i := 0; i < numberOfWorkers; i++ {
		go probeWorker(k, jobsChan, results)
	}
	var froms, tos []string
	fromSet := map[string]bool{}
	toSet := map[string]bool{}
	for i, job := range jobs {
		log.Infof("queueing up probe job %d", i+1)
		jobsChan <- job
		if _, ok := fromSet[job.GetFromKey()]; !ok {
			froms = append(froms, job.GetFromKey())
			fromSet[job.GetFromKey()] = true
		}
		if _, ok := toSet[job.GetToKey()]; !ok {
			tos = append(tos, job.GetToKey())
			toSet[job.GetToKey()] = true
		}
	}
	close(jobsChan)

	table := netpol.NewStringTruthTableWithFromsTo(froms, tos)

	for i := 0; i < len(jobs); i++ {
		log.Debugf("handling results from probe job %d", i+1)
		result := <-results
		job := result.Job
		var entry string
		if result.Result.Err != "" {
			log.Infof("unable to perform probe %s/%s/%s -> %s:%d : %v", job.FromNamespace, job.FromPod, job.FromContainer, job.ToAddress, job.ToPort, result.Result.Err)
			entry = fmt.Sprintf("error: %d", result.Result.ExitCode)
		} else {
			entry = fmt.Sprintf("%d", result.Result.ExitCode)
		}
		table.Set(job.GetFromKey(), job.GetToKey(), entry)
	}
	return table
}

func probeWorker(k8s *Kubernetes, jobs <-chan *ProbeJob, results chan<- *ProbeJobResult) {
	for job := range jobs {
		log.Infof("starting probe job %+v", job)
		result, err := k8s.Probe(job)
		jobResult := &ProbeJobResult{
			Job:    job,
			Result: result,
			Err:    err,
		}
		log.Infof("finished probe job %+v", result)
		results <- jobResult
	}
}

// convenience functions

func (k *Kubernetes) ProbePodToPod(namespaces []string, timeoutSeconds int) (*netpol.StringTruthTable, error) {
	pods, err := k.GetPodsInNamespaces(namespaces)
	if err != nil {
		return nil, err
	}

	var jobs []*ProbeJob
	// TODO could verify that each pod only has one container, each container only has one port, etc.
	for _, from := range pods {
		for _, to := range pods {
			jobs = append(jobs, &ProbeJob{
				FromNamespace:  from.Namespace,
				FromPod:        from.Name,
				FromContainer:  from.Spec.Containers[0].Name,
				ToAddress:      to.Status.PodIP,
				ToPort:         int(to.Spec.Containers[0].Ports[0].ContainerPort),
				TimeoutSeconds: timeoutSeconds,
				// TODO
				//FromKey:            "",
				//ToKey:              "",
			})
		}
	}

	return k.ProbeConnectivity(jobs), nil
}
