package probe

import (
	"time"

	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Resources struct {
	Namespaces map[string]map[string]string
	Pods       []*Pod
	Nodes      []*Node
	Services   map[string]*v1.Service
	//ExternalIPs []string
}

func NewDefaultResources(kubernetes kube.IKubernetes, namespaces []string, podNames []string, ports []int, protocols []v1.Protocol, externalIPs []string, podCreationTimeoutSeconds int, batchJobs bool) (*Resources, error) {
	//sort.Strings(externalIPs) // TODO why is this here?

	r := &Resources{
		Namespaces: map[string]map[string]string{},
		Services:   make(map[string]*v1.Service),

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
	if err := r.getNamespaceLabelsFromKube(kubernetes); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Resources) waitForPodsReady(kubernetes kube.IKubernetes, timeoutSeconds int) error {
	sleep := 5
	for i := 0; i < timeoutSeconds; i += sleep {
		podList, err := kube.GetPodsInNamespaces(kubernetes, r.NamespacesSlice())
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
	podList, err := kube.GetPodsInNamespaces(kubernetes, r.NamespacesSlice())
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
		kubeService, err := kubernetes.GetService(pod.Namespace, pod.ServiceName())
		if err != nil {
			return err
		}
		pod.ServiceIP = kubeService.Spec.ClusterIP

		logrus.Debugf("ip for pod %s/%s: %s", pod.Namespace, pod.Name, pod.IP)
	}

	return nil
}

func (r *Resources) getNamespaceLabelsFromKube(kubernetes kube.IKubernetes) error {
	nsList, err := kubernetes.GetAllNamespaces()
	if err != nil {
		return err
	}

	for _, kubeNs := range nsList.Items {
		for label, value := range kubeNs.Labels {
			if ns, ok := r.Namespaces[kubeNs.Name]; ok {
				ns[label] = value
			}
		}
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
		Nodes:      r.Nodes,
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
		Nodes:      r.Nodes,
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
		Nodes:      r.Nodes,
	}, nil
}

// CreateServce returns a new object with a new namespace.  It should not affect the original Resources object.
func (r *Resources) CreateService(svc *v1.Service) (*Resources, error) {
	if _, ok := r.Services[svc.Name]; ok {
		return nil, errors.Errorf("service %s already found", svc.Name)
	}
	newServices := map[string]*v1.Service{}
	for oldServiceName, oldService := range r.Services {
		newServices[oldServiceName] = oldService // Note: service type is pointer, duplicate resource type needs to be deep copied
	}
	newServices[svc.Name] = svc
	return &Resources{
		Services:   newServices,
		Pods:       r.Pods,
		Nodes:      r.Nodes,
		Namespaces: r.Namespaces,
	}, nil
}

// DeleteNamespace returns a new object without the namespace.  It should not affect the original Resources object.
func (r *Resources) DeleteService(svc *v1.Service) (*Resources, error) {
	if _, ok := r.Services[svc.Name]; !ok {
		return nil, errors.Errorf("service %s/%s not found in test state", svc.Namespace, svc.Name)
	}
	newServices := map[string]*v1.Service{}
	for oldServiceName, oldService := range r.Services {
		if oldServiceName != svc.Name {
			newServices[oldServiceName] = oldService
		}
	}

	return &Resources{
		Services: newServices,
	}, nil
}

func (r *Resources) addNodes(nodes *v1.NodeList) {
	for _, node := range nodes.Items {
		nodeips := node.Status.Addresses
		if len(nodeips) > 0 {
			logrus.Debugf("loading node name %s and ip %+v", node.Name, nodeips[0].Address)
			r.Nodes = append(r.Nodes, NewNode(node.Name, node.Labels, nodeips[0].Address))
		} else {
			logrus.Errorf("node %s has no ip's", node.Name)
		}
	}
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
		Nodes:      r.Nodes,
		Pods:       append(append([]*Pod{}, r.Pods...), NewPod(ns, podName, labels, "TODO", r.Pods[0].Containers)),
		//ExternalIPs: r.ExternalIPs,
	}, nil
}

// SetPodLabels returns a new object with an updated pod.  It should not affect the original Resources object.
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
		Nodes:      r.Nodes,
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
		Nodes:      r.Nodes,
		//ExternalIPs: r.ExternalIPs,
	}, nil
}

func (r *Resources) SortedPodNames() []string {
	return slice.Sort(slice.Map(
		func(p *Pod) string { return p.PodString().String() },
		r.Pods))
}

func (r *Resources) SortedNodeNames() []string {
	return slice.Sort(slice.Map(
		func(n *Node) string { return n.Name },
		r.Nodes))
}

func (r *Resources) NamespacesSlice() []string {
	return maps.Keys(r.Namespaces)
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
		kubeServiceLoadBalancer := pod.KubeServiceLoadBalancer()
		_, err = kubernetes.GetService(kubeService.Namespace, kubeService.Name)
		if err != nil {
			_, err = kubernetes.CreateService(kubeService)
			if err != nil {
				return err
			}
			_, err = kubernetes.CreateService(kubeServiceLoadBalancer)
			if err != nil {
				return err
			}
		}
	}

	nodes, err := kubernetes.GetNodes()
	if err != nil {
		return err
	}
	r.addNodes(nodes)

	return err
}

func KubeNamespace(ns string, labels map[string]string) *v1.Namespace {
	return &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns, Labels: labels}}
}
