package netpol

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

type Pod string

func NewPod(namespace string, podName string) Pod {
	return Pod(fmt.Sprintf("%s/%s", namespace, podName))
}

func (pod Pod) split() (string, string) {
	pieces := strings.Split(string(pod), "/")
	if len(pieces) != 2 {
		panic(errors.New(fmt.Sprintf("expected ns/pod, found %+v", pieces)))
	}
	return pieces[0], pieces[1]
}

func (pod Pod) Namespace() string {
	ns, _ := pod.split()
	return ns
}

func (pod Pod) PodName() string {
	_, podName := pod.split()
	return podName
}
