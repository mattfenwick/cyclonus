package kube

import "fmt"

func QualifiedServiceAddress(ns string, service string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", service, ns)
}
