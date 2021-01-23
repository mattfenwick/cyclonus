package connectivity

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const (
	agnhostImage = "k8s.gcr.io/e2e-test-images/agnhost:2.21"
)

func CreateResources(kube *kube.Kubernetes, model *PodModel) error {
	for nsName, ns := range model.Namespaces {
		_, err := kube.CreateOrUpdateNamespace(NamespaceSpec(nsName, ns.Labels))
		if err != nil {
			return err
		}
		for podName, pod := range ns.Pods {
			kubePod := KubePod(nsName, podName, pod.Labels, pod.Containers)
			_, err = kube.CreatePodIfNotExists(kubePod)
			if err != nil {
				return err
			}
			kubeService := Service(nsName, podName, pod.Labels, pod.Containers)
			_, err = kube.CreateServiceIfNotExists(kubeService)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func NamespaceSpec(name string, labels map[string]string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

// KubePod returns the kube pod
func KubePod(namespace string, name string, labels map[string]string, containers []*Container) *v1.Pod {
	zero := int64(0)
	var kubeContainers []v1.Container
	for _, cont := range containers {
		kubeContainers = append(kubeContainers, KubeContainer(cont.Name, cont.Protocol, cont.Port))
	}
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Labels:    labels,
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			TerminationGracePeriodSeconds: &zero,
			Containers:                    kubeContainers,
		},
	}
}

// Service returns a kube service spec
func Service(namespace string, pod string, labels map[string]string, containers []*Container) *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName(namespace, pod),
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Selector: labels,
		},
	}
	for _, cont := range containers {
		service.Spec.Ports = append(service.Spec.Ports, v1.ServicePort{
			Name:     fmt.Sprintf("service-port-%s-%d", strings.ToLower(string(cont.Protocol)), cont.Port),
			Protocol: cont.Protocol,
			Port:     int32(cont.Port),
		})
	}
	return service
}

func KubeContainer(name string, protocol v1.Protocol, port int) v1.Container {
	var cmd []string

	switch protocol {
	case v1.ProtocolTCP:
		cmd = []string{"/agnhost", "serve-hostname", "--tcp", "--http=false", "--port", fmt.Sprintf("%d", port)}
	case v1.ProtocolUDP:
		cmd = []string{"/agnhost", "serve-hostname", "--udp", "--http=false", "--port", fmt.Sprintf("%d", port)}
	case v1.ProtocolSCTP:
		cmd = []string{"/agnhost", "netexec", "--sctp-port", fmt.Sprintf("%d", port)}
	default:
		panic(errors.Errorf("invalid protocol %s", protocol))
	}
	return v1.Container{
		Name:            name,
		ImagePullPolicy: v1.PullIfNotPresent,
		Image:           agnhostImage,
		Command:         cmd,
		SecurityContext: &v1.SecurityContext{},
		Ports: []v1.ContainerPort{
			{
				ContainerPort: int32(port),
				Name:          fmt.Sprintf("serve-%d-%s", port, strings.ToLower(string(protocol))),
				Protocol:      protocol,
			},
		},
	}
}

//

type KubeProbeJob struct {
	FromPod  *NamespacedPod
	ToPod    *NamespacedPod
	Port     int
	Protocol v1.Protocol
}

func (pj *KubeProbeJob) ToAddress() string {
	return kube.QualifiedServiceAddress(pj.ToPod.ServiceName(), pj.ToPod.NamespaceName)
}

func (pj *KubeProbeJob) FromContainer() string {
	return pj.FromPod.Containers[0].Name
}

func (pj *KubeProbeJob) ClientCommand() []string {
	switch pj.Protocol {
	case v1.ProtocolSCTP:
		return []string{"/agnhost", "connect", fmt.Sprintf("%s:%d", pj.ToAddress(), pj.Port), "--timeout=1s", "--protocol=sctp"}
	case v1.ProtocolTCP:
		return []string{"/agnhost", "connect", fmt.Sprintf("%s:%d", pj.ToAddress(), pj.Port), "--timeout=1s", "--protocol=tcp"}
	case v1.ProtocolUDP:
		return []string{"nc", "-v", "-z", "-w", "1", "-u", pj.ToAddress(), fmt.Sprintf("%d", pj.Port)}
	default:
		panic(errors.Errorf("protocol %s not supported", pj.Protocol))
	}
}

func (pj *KubeProbeJob) KubeExecCommand() []string {
	return append([]string{
		"kubectl", "exec",
		pj.FromPod.PodName,
		"-c", pj.FromContainer(),
		"-n", pj.FromPod.NamespaceName,
		"--",
	},
		pj.ClientCommand()...)
}

func (pj *KubeProbeJob) ToURL() string {
	return fmt.Sprintf("http://%s:%d", pj.ToAddress(), pj.Port)
}

type KubeProbeJobResults struct {
	Job         *KubeProbeJob
	IsConnected bool
	Err         error
	Command     string
}

func RunKubeProbe(k8s *kube.Kubernetes, model *PodModel, port int, protocol v1.Protocol, numberOfWorkers int) *TruthTable {
	allPods := model.AllPods()
	size := len(allPods) * len(allPods)
	jobs := make(chan *KubeProbeJob, size)
	results := make(chan *KubeProbeJobResults, size)
	for i := 0; i < numberOfWorkers; i++ {
		go probeWorker(k8s, jobs, results)
	}
	for _, podFrom := range allPods {
		for _, podTo := range allPods {
			jobs <- &KubeProbeJob{
				FromPod:  podFrom,
				ToPod:    podTo,
				Protocol: protocol,
				Port:     port,
			}
		}
	}
	close(jobs)

	reachability := model.NewTruthTable()
	for i := 0; i < size; i++ {
		result := <-results
		job := result.Job
		if result.Err != nil {
			log.Errorf("unable to perform probe %s -> %s: %v", job.FromPod.PodString().String(), job.ToAddress(), result.Err)
		}
		reachability.Set(job.FromPod.PodString().String(), job.ToPod.PodString().String(), result.IsConnected)
	}
	return reachability
}

// probeWorker continues polling a pod connectivity status, until the incoming "jobs" channel is closed, and writes results back out to the "results" channel.
// it only writes pass/fail status to a channel and has no failure side effects, this is by design since we do not want to fail inside a goroutine.
func probeWorker(k8s *kube.Kubernetes, jobs <-chan *KubeProbeJob, results chan<- *KubeProbeJobResults) {
	for job := range jobs {
		connected, command, err := probeConnectivity(k8s, job)
		result := &KubeProbeJobResults{
			Job:         job,
			IsConnected: connected,
			Err:         err,
			Command:     command,
		}
		results <- result
	}
}

func probeConnectivity(k8s *kube.Kubernetes, job *KubeProbeJob) (bool, string, error) {
	commandDebugString := strings.Join(job.KubeExecCommand(), " ")
	stdout, stderr, commandErr, err := k8s.ExecuteRemoteCommand(job.FromPod.NamespaceName, job.FromPod.PodName, job.FromContainer(), job.ClientCommand())
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
