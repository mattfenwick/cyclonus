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

func NewDefaultResources(kubernetes *kube.Kubernetes, namespaces []string, podNames []string, ports []int, protocols []v1.Protocol, externalIPs []string, podCreationTimeoutSeconds int) (*Resources, error) {
	sort.Strings(externalIPs)
	r := &Resources{
		Namespaces: map[string]map[string]string{},
		//ExternalIPs: externalIPs,
	}

	for _, ns := range namespaces {
		for _, podName := range podNames {
			r.Pods = append(r.Pods, NewDefaultPod(ns, podName, map[string]string{"pod": podName}, "TODO", ports, protocols))
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

func (r *Resources) getPodIPsFromKube(kubernetes *kube.Kubernetes) error {
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

// UpdateNamespaceLabels returns a new object with an updated namespace.  It should not affect the original Resources object.
func (r *Resources) UpdateNamespaceLabels(ns string, labels map[string]string) (*Resources, error) {
	if _, ok := r.Namespaces[ns]; !ok {
		return nil, errors.Errorf("no namespace %s", ns)
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

func (r *Resources) CreateResourcesInKube(kubernetes *kube.Kubernetes) error {
	for ns, labels := range r.Namespaces {
		_, err := kubernetes.CreateOrUpdateNamespace(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns, Labels: labels}})
		if err != nil {
			return err
		}
	}
	for _, pod := range r.Pods {
		_, err := kubernetes.CreatePodIfNotExists(pod.KubePod())
		if err != nil {
			return err
		}
		_, err = kubernetes.CreateServiceIfNotExists(pod.KubeService())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Resources) VerifyClusterState(kubernetes *kube.Kubernetes) error {
	kubePods, err := kubernetes.GetPodsInNamespaces(r.NamespacesSlice())
	if err != nil {
		return err
	}

	// 1. pods: labels, ips, containers, ports
	actualPods := map[string]v1.Pod{}
	for _, kubePod := range kubePods {
		actualPods[NewPodString(kubePod.Namespace, kubePod.Name).String()] = kubePod
	}
	// are we missing any pods?
	for _, pod := range r.Pods {
		if actualPod, ok := actualPods[pod.PodString().String()]; ok {
			if !areLabelsEqual(actualPod.Labels, pod.Labels) {
				return errors.Errorf("for pod %s, expected labels %+v (found %+v)", pod.PodString().String(), pod.Labels, actualPod.Labels)
			}
			if actualPod.Status.PodIP != pod.IP {
				return errors.Errorf("for pod %s, expected ip %s (found %s)", pod.PodString().String(), pod.IP, actualPod.Status.PodIP)
			}
			if !areContainersEqual(actualPod, pod) {
				return errors.Errorf("for pod %s, expected containers %+v (found %+v)", pod.PodString().String(), pod.Containers, actualPod.Spec.Containers)
			}
		} else {
			return errors.Errorf("missing expected pod %s", pod.PodString().String())
		}
	}

	// 2. services: selectors, ports
	for _, pod := range r.Pods {
		expected := pod.KubeService()
		svc, err := kubernetes.GetService(expected.Namespace, expected.Name)
		if err != nil {
			return err
		}
		if !areLabelsEqual(svc.Spec.Selector, pod.Labels) {
			return errors.Errorf("for service %s/%s, expected labels %+v (found %+v)", pod.Namespace, pod.Name, pod.Labels, svc.Spec.Selector)
		}
		if len(expected.Spec.Ports) != len(svc.Spec.Ports) {
			return errors.Errorf("for service %s/%s, expected %d ports (found %d)", expected.Namespace, expected.Name, len(expected.Spec.Ports), len(svc.Spec.Ports))
		}
		for i, port := range expected.Spec.Ports {
			kubePort := svc.Spec.Ports[i]
			if kubePort.Protocol != port.Protocol || kubePort.Port != port.Port {
				return errors.Errorf("for service %s/%s, expected port %+v (found %+v)", expected.Namespace, expected.Name, port, kubePort)
			}
		}
	}

	// 3. namespaces: names, labels
	for ns, labels := range r.Namespaces {
		namespace, err := kubernetes.GetNamespace(ns)
		if err != nil {
			return err
		}
		if !areLabelsEqual(namespace.Labels, labels) {
			return errors.Errorf("for namespace %s, expected labels %+v (found %+v)", ns, labels, namespace.Labels)
		}
	}

	// nothing wrong: we're good to go
	return nil
}

func areContainersEqual(kubePod v1.Pod, expectedPod *Pod) bool {
	kubeConts := kubePod.Spec.Containers
	if len(kubeConts) != len(expectedPod.Containers) {
		return false
	}
	for i, kubeCont := range kubeConts {
		cont := expectedPod.Containers[i]
		if len(kubeCont.Ports) != 1 {
			return false
		}
		if int(kubeCont.Ports[0].ContainerPort) != cont.Port {
			return false
		}
		if kubeCont.Ports[0].Protocol != cont.Protocol {
			return false
		}
	}

	return true
}

func areLabelsEqual(l map[string]string, r map[string]string) bool {
	if len(l) != len(r) {
		return false
	}
	for k, lv := range l {
		rv, ok := r[k]
		if !ok || lv != rv {
			return false
		}
	}
	return true
}

func (r *Resources) ResetLabelsInKube(kubernetes *kube.Kubernetes) error {
	for ns, labels := range r.Namespaces {
		_, err := kubernetes.SetNamespaceLabels(ns, labels)
		if err != nil {
			return err
		}
	}

	for _, pod := range r.Pods {
		_, err := kubernetes.SetPodLabels(pod.Namespace, pod.Name, pod.Labels)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Resources) waitForPodsReady(kubernetes *kube.Kubernetes, timeoutSeconds int) error {
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

		logrus.Infof("waiting for pods to be running and have IP addresses")
		time.Sleep(time.Duration(sleep) * time.Second)
	}
	return errors.Errorf("pods not ready")
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

func (r *Resources) AllProtocolsServed() map[v1.Protocol]bool {
	protocols := map[v1.Protocol]bool{}
	for _, pod := range r.Pods {
		for _, cont := range pod.Containers {
			protocols[cont.Protocol] = true
		}
	}
	return protocols
}
