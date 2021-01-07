package kube

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	v1net "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

type Kubernetes struct {
	podCache  map[string][]v1.Pod
	ClientSet *kubernetes.Clientset
}

func NewKubernetes() (*Kubernetes, error) {
	clientSet, err := Clientset()
	if err != nil {
		return nil, err
	}
	return &Kubernetes{
		podCache:  map[string][]v1.Pod{},
		ClientSet: clientSet,
	}, nil
}

func Clientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := filepath.Join(
			os.Getenv("HOME"), ".kube", "config",
		)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to build config from flags, check that your KUBECONFIG file is correct !")
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to instantiate clientset")
	}
	return clientset, nil
}

//func (k *Kubernetes) GetPods(ns string, key string, val string) ([]v1.Pod, error) {
//	if p, ok := k.podCache[fmt.Sprintf("%v_%v_%v", ns, key, val)]; ok {
//		return p, nil
//	}
//
//	v1PodList, err := k.ClientSet.CoreV1().Pods(ns).List(metav1.ListOptions{})
//	if err != nil {
//		return nil, errors.Wrapf(err, "unable to list pods")
//	}
//	pods := []v1.Pod{}
//	for _, pod := range v1PodList.Items {
//		// log.Infof("check: %s, %s, %s, %s", pod.Name, pod.Labels, key, val)
//		if pod.Labels[key] == val {
//			pods = append(pods, pod)
//		}
//	}
//
//	//log.Infof("list in ns %s: %d -> %d", ns, len(v1PodList.Items), len(pods))
//	k.podCache[fmt.Sprintf("%v_%v_%v", ns, key, val)] = pods
//
//	return pods, nil
//}

//func (k *Kubernetes) GetPod(ns string, key string, val string) (v1.Pod, error) {
//	pods := k.GetPods(ns, key, val)
//	if len(pods) != 1 {
//		return errors.Errorf("expected 1 pod of ")
//	}
//}

func (k *Kubernetes) CreateOrUpdateNamespace(n string, labels map[string]string) (*v1.Namespace, error) {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   n,
			Labels: labels,
		},
	}
	nsr, err := k.ClientSet.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	if err == nil {
		log.Infof("created namespace %s", ns)
		return nsr, errors.Wrapf(err, "unable to create namespace %s", ns)
	}

	log.Debugf("unable to create namespace %s, let's try updating it instead (error: %s)", ns.Name, err)
	nsr, err = k.ClientSet.CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{})
	if err != nil {
		log.Debugf("unable to create namespace %s: %s", ns, err)
	}

	return nsr, err
}

func (k *Kubernetes) CleanNetworkPolicies(ns string) error {
	netpols, err := k.ClientSet.NetworkingV1().NetworkPolicies(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "unable to list network policies in ns %s", ns)
	}
	for _, np := range netpols.Items {
		log.Infof("deleting network policy %s/%s", ns, np.Name)
		err = k.ClientSet.NetworkingV1().NetworkPolicies(np.Namespace).Delete(context.TODO(), np.Name, metav1.DeleteOptions{})
		if err != nil {
			return errors.Wrapf(err, "unable to delete netpol %s/%s", ns, np.Name)
		}
	}
	return nil
}

func (k *Kubernetes) CreateNetworkPolicy(netpol *v1net.NetworkPolicy) (*v1net.NetworkPolicy, error) {
	ns := netpol.Namespace
	log.Infof("creating network policy %s in ns %s", netpol.Name, ns)

	createdPolicy, err := k.ClientSet.NetworkingV1().NetworkPolicies(ns).Create(context.TODO(), netpol, metav1.CreateOptions{})
	return createdPolicy, errors.Wrapf(err, "unable to create network policy %s/%s", netpol.Name, netpol.Namespace)
}

func (k *Kubernetes) CreateOrUpdateNetworkPolicy(ns string, netpol *v1net.NetworkPolicy) (*v1net.NetworkPolicy, error) {
	log.Infof("creating/updating network policy %s/%s", ns, netpol.Name)
	netpol.ObjectMeta.Namespace = ns
	np, err := k.ClientSet.NetworkingV1().NetworkPolicies(ns).Update(context.TODO(), netpol, metav1.UpdateOptions{})
	if err == nil {
		return np, err
	}

	log.Debugf("unable to update network policy %s/%s, let's try creating it instead (error: %s)", ns, netpol.Name, err)
	np, err = k.ClientSet.NetworkingV1().NetworkPolicies(ns).Create(context.TODO(), netpol, metav1.CreateOptions{})
	if err != nil {
		log.Debugf("unable to create network policy %s/%s: %s", ns, netpol.Name, err)
	}
	return np, err
}

func (k *Kubernetes) CreateDaemonSet(namespace string, ds *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	return k.ClientSet.AppsV1().DaemonSets(namespace).Create(context.TODO(), ds, metav1.CreateOptions{})
}

func (k *Kubernetes) CreateDaemonSetIfNotExists(namespace string, ds *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	created, err := k.ClientSet.AppsV1().DaemonSets(namespace).Create(context.TODO(), ds, metav1.CreateOptions{})
	if err == nil {
		return created, nil
	}
	if err.Error() == fmt.Sprintf(`daemonsets.apps "%s" already exists`, ds.Name) {
		return nil, nil
	}
	return nil, err
}

func (k *Kubernetes) CreateService(namespace string, svc *v1.Service) (*v1.Service, error) {
	return k.ClientSet.CoreV1().Services(namespace).Create(context.TODO(), svc, metav1.CreateOptions{})
}

func (k *Kubernetes) CreateServiceIfNotExists(namespace string, svc *v1.Service) (*v1.Service, error) {
	created, err := k.ClientSet.CoreV1().Services(namespace).Create(context.TODO(), svc, metav1.CreateOptions{})
	if err == nil {
		return created, nil
	}
	if err.Error() == fmt.Sprintf(`services "%s" already exists`, svc.Name) {
		return nil, nil
	}
	return nil, err
}

func (k *Kubernetes) GetPodsInNamespaces(namespaces []string) ([]v1.Pod, error) {
	var pods []v1.Pod
	for _, ns := range namespaces {
		podList, err := k.ClientSet.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get pods in namespace %s", ns)
		}
		pods = append(pods, podList.Items...)
	}
	return pods, nil
}
