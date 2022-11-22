package kube

import "fmt"

// QualifiedServiceAddress returns the address that can be used to hit a service from
// any namespace in the cluster
//
//	func QualifiedServiceAddress(serviceName string, namespace string, dnsDomain string) string {
//		return fmt.Sprintf("%s.%s.svc.%s", serviceName, namespace, dnsDomain)
func QualifiedServiceAddress(serviceName string, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace)
}
