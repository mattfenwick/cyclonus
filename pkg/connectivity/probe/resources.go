package probe

import (
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sort"
	"time"
)

type Resources struct {
	Namespaces map[string]map[string]string
	Pods       []*Pod
	//ExternalIPs []string
}

func NewDefaultResources(kubernetes kube.IKubernetes, namespaces []string, podNames []string, ports []int, protocols []v1.Protocol, externalIPs []string, podCreationTimeoutSeconds int, batchJobs bool) (*Resources, error) {
	sort.Strings(externalIPs)
	r := &Resources{
		Namespaces: map[string]map[string]string{},
		//ExternalIPs: externalIPs,
	}

	for _, ns := range namespaces {
		for _, podName := range podNames {
			r.Pods = append(r.Pods, NewDefaultPod(ns, podName, ports, protocols, batchJobs))
		}
		r.Namespaces[ns] = map[string]string{"ns": ns}
	}

	if err := r.CreateResourcesInKube(kubernetes); err != nil {
		return nil, err
	}
	if err := r.waitForPodsReady(kubernetes, podCreationTimeoutSeconds); err != nil {
		return nil, err
	}
	if err := r.getPodIPsFromKube(kubernetes); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Resources) waitForPodsReady(kubernetes kube.IKubernetes, timeoutSeconds int) error {
	sleep := 5
	for i := 0; i < timeoutSeconds; i += sleep {
		podList, err := kubernetes.GetPodsInNamespaces(r.NamespacesSlice())
		if err != nil {
			return err
		}

		ready := 0
		for _, pod := range podList {
			if pod.Status.Phase == "Running" && pod.Status.PodIP != "" {
				ready++
			}
		}
		if ready == len(r.Pods) {
			return nil
		}

		logrus.Infof("waiting for %d pods to be running and have IP addresses; currently %d are ready", len(r.Pods), ready)
		time.Sleep(time.Duration(sleep) * time.Second)
	}
	return errors.Errorf("pods not ready")
}

func (r *Resources) getPodIPsFromKube(kubernetes kube.IKubernetes) error {
	podList, err := kubernetes.GetPodsInNamespaces(r.NamespacesSlice())
	if err != nil {
		return err
	}

	for _, kubePod := range podList {
		if kubePod.Status.PodIP == "" {
			return errors.Errorf("no ip found for pod %s/%s", kubePod.Namespace, kubePod.Name)
		}

		pod, err := r.GetPod(kubePod.Namespace, kubePod.Name)
		if err != nil {
			return errors.Errorf("unable to find pod %s/%s in resources", kubePod.Namespace, kubePod.Name)
		}
		pod.IP = kubePod.Status.PodIP

		logrus.Debugf("ip for pod %s/%s: %s", pod.Namespace, pod.Name, pod.IP)
	}

	return nil
}

func (r *Resources) GetPod(ns string, name string) (*Pod, error) {
	for _, pod := range r.Pods {
		if pod.Namespace == ns && pod.Name == name {
			return pod, nil
		}
	}
	return nil, errors.Errorf("unable to find pod %s/%s", ns, name)
}

// CreateNamespace returns a new object with a new namespace.  It should not affect the original Resources object.
func (r *Resources) CreateNamespace(ns string, labels map[string]string) (*Resources, error) {
	if _, ok := r.Namespaces[ns]; ok {
		return nil, errors.Errorf("namespace %s already found", ns)
	}
	newNamespaces := map[string]map[string]string{}
	for oldNs, oldLabels := range r.Namespaces {
		newNamespaces[oldNs] = oldLabels
	}
	newNamespaces[ns] = labels
	return &Resources{
		Namespaces: newNamespaces,
		Pods:       r.Pods,
	}, nil
}

// UpdateNamespaceLabels returns a new object with an updated namespace.  It should not affect the original Resources object.
func (r *Resources) UpdateNamespaceLabels(ns string, labels map[string]string) (*Resources, error) {
	if _, ok := r.Namespaces[ns]; !ok {
		return nil, errors.Errorf("namespace %s not found", ns)
	}
	newNamespaces := map[string]map[string]string{}
	for oldNs, oldLabels := range r.Namespaces {
		newNamespaces[oldNs] = oldLabels
	}
	newNamespaces[ns] = labels
	return &Resources{
		Namespaces: newNamespaces,
		Pods:       r.Pods,
	}, nil
}

// DeleteNamespace returns a new object without the namespace.  It should not affect the original Resources object.
func (r *Resources) DeleteNamespace(ns string) (*Resources, error) {
	if _, ok := r.Namespaces[ns]; !ok {
		return nil, errors.Errorf("namespace %s not found", ns)
	}
	newNamespaces := map[string]map[string]string{}
	for oldNs, oldLabels := range r.Namespaces {
		if oldNs != ns {
			newNamespaces[oldNs] = oldLabels
		}
	}
	var pods []*Pod
	for _, pod := range r.Pods {
		if pod.Namespace == ns {
			// skip
		} else {
			pods = append(pods, pod)
		}
	}
	return &Resources{
		Namespaces: newNamespaces,
		Pods:       pods,
	}, nil
}

// CreatePod returns a new object with a new pod.  It should not affect the original Resources object.
func (r *Resources) CreatePod(ns string, podName string, labels map[string]string) (*Resources, error) {
	// TODO this needs to be improved
	//   for now, let's assume all pods have the same containers and just copy the containers from the first pod
	if _, ok := r.Namespaces[ns]; !ok {
		return nil, errors.Errorf("can't find namespace %s", ns)
	}
	return &Resources{
		Namespaces: r.Namespaces,
		Pods:       append(append([]*Pod{}, r.Pods...), NewPod(ns, podName, labels, "TODO", r.Pods[0].Containers)),
		//ExternalIPs: r.ExternalIPs,
	}, nil
}

// UpdatePodLabel returns a new object with an updated pod.  It should not affect the original Resources object.
func (r *Resources) SetPodLabels(ns string, podName string, labels map[string]string) (*Resources, error) {
	var pods []*Pod
	found := false
	for _, existingPod := range r.Pods {
		if existingPod.Namespace == ns && existingPod.Name == podName {
			found = true
			pods = append(pods, existingPod.SetLabels(labels))
		} else {
			pods = append(pods, existingPod)
		}
	}
	if !found {
		return nil, errors.Errorf("no pod named %s/%s found", ns, podName)
	}
	return &Resources{
		Namespaces: r.Namespaces,
		Pods:       pods,
		//ExternalIPs: r.ExternalIPs,
	}, nil
}

// DeletePod returns a new object without the deleted pod.  It should not affect the original Resources object.
func (r *Resources) DeletePod(ns string, podName string) (*Resources, error) {
	var newPods []*Pod
	found := false
	for _, pod := range r.Pods {
		if pod.Namespace == ns && pod.Name == podName {
			found = true
		} else {
			newPods = append(newPods, pod)
		}
	}
	if !found {
		return nil, errors.Errorf("pod %s/%s not found", ns, podName)
	}
	return &Resources{
		Namespaces: r.Namespaces,
		Pods:       newPods,
		//ExternalIPs: r.ExternalIPs,
	}, nil
}

func (r *Resources) SortedPodNames() []string {
	var podNames []string
	for _, pod := range r.Pods {
		podNames = append(podNames, pod.PodString().String())
	}
	sort.Strings(podNames)
	return podNames
}

func (r *Resources) NamespacesSlice() []string {
	var nss []string
	for ns := range r.Namespaces {
		nss = append(nss, ns)
	}
	return nss
}

func (r *Resources) CreateResourcesInKube(kubernetes kube.IKubernetes) error {
	for ns, labels := range r.Namespaces {
		_, err := kubernetes.GetNamespace(ns)
		if err != nil {
			_, err := kubernetes.CreateNamespace(KubeNamespace(ns, labels))
			if err != nil {
				return err
			}
		}
	}
	for _, pod := range r.Pods {
		_, err := kubernetes.GetPod(pod.Namespace, pod.Name)
		if err != nil {
			_, err := kubernetes.CreatePod(pod.KubePod())
			if err != nil {
				return err
			}
		}
		kubeService := pod.KubeService()
		_, err = kubernetes.GetService(kubeService.Namespace, kubeService.Name)
		if err != nil {
			_, err = kubernetes.CreateService(kubeService)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func KubeNamespace(ns string, labels map[string]string) *v1.Namespace {
	return &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns, Labels: labels}}
}

func (r *Resources) GetJobsForNamedPortProtocol(port intstr.IntOrString, protocol v1.Protocol) *Jobs {
	jobs := &Jobs{}
	for _, podFrom := range r.Pods {
		for _, podTo := range r.Pods {
			job := &Job{
				FromKey:             podFrom.PodString().String(),
				FromNamespace:       podFrom.Namespace,
				FromNamespaceLabels: r.Namespaces[podFrom.Namespace],
				FromPod:             podFrom.Name,
				FromPodLabels:       podFrom.Labels,
				FromContainer:       podFrom.Containers[0].Name,
				FromIP:              podFrom.IP,
				ToKey:               podTo.PodString().String(),
				ToHost:              kube.QualifiedServiceAddress(podTo.ServiceName(), podTo.Namespace),
				ToNamespace:         podTo.Namespace,
				ToNamespaceLabels:   r.Namespaces[podTo.Namespace],
				ToPodLabels:         podTo.Labels,
				ToIP:                podTo.IP,
				ResolvedPort:        -1,
				ResolvedPortName:    "",
				Protocol:            protocol,
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

func (r *Resources) GetJobsAllAvailableServers() *Jobs {
	var jobs []*Job
	for _, podFrom := range r.Pods {
		for _, podTo := range r.Pods {
			for _, contTo := range podTo.Containers {
				jobs = append(jobs, &Job{
					FromKey:             podFrom.PodString().String(),
					FromNamespace:       podFrom.Namespace,
					FromNamespaceLabels: r.Namespaces[podFrom.Namespace],
					FromPod:             podFrom.Name,
					FromPodLabels:       podFrom.Labels,
					FromContainer:       podFrom.Containers[0].Name,
					FromIP:              podFrom.IP,
					ToKey:               podTo.PodString().String(),
					ToHost:              kube.QualifiedServiceAddress(podTo.ServiceName(), podTo.Namespace),
					ToNamespace:         podTo.Namespace,
					ToNamespaceLabels:   r.Namespaces[podTo.Namespace],
					ToPodLabels:         podTo.Labels,
					ToContainer:         contTo.Name,
					ToIP:                podTo.IP,
					ResolvedPort:        contTo.Port,
					ResolvedPortName:    contTo.PortName,
					Protocol:            contTo.Protocol,
				})
			}
		}
	}
	return &Jobs{Valid: jobs}
}
