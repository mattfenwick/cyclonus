package synthetic

import (
	"github.com/mattfenwick/cyclonus/pkg/connectivity/types"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func RunSyntheticProbe(request *Request) *Result {
	resources := request.Resources
	resultTable := resources.NewResultTable()

	log.Infof("running synthetic probe on port %s, protocol %s", request.Port.String(), request.Protocol)

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

			hasServer := false
			// TODO resolve to container, or fail appropriately
			for _, c := range podTo.Containers {
				switch request.Port.Type {
				case intstr.Int:
					if c.Port == int(request.Port.IntVal) {
						hasServer = true
					}
				case intstr.String:
					if c.PortName == request.Port.StrVal {
						hasServer = true
					}
				}
			}

			fr := podFrom.PodString().String()
			to := podTo.PodString().String()
			allowed := request.Policies.IsTrafficAllowed(traffic)
			// TODO could also keep the whole `allowed` struct somewhere
			if hasServer {
				if allowed.Ingress.IsAllowed() {
					resultTable.SetIngress(fr, to, types.ConnectivityAllowed)
				} else {
					resultTable.SetIngress(fr, to, types.ConnectivityBlocked)
				}
			} else {
				resultTable.SetIngress(fr, to, types.ConnectivityInvalidPortProtocol)
			}
			if allowed.Egress.IsAllowed() {
				resultTable.SetEgress(fr, to, types.ConnectivityAllowed)
			} else {
				resultTable.SetEgress(fr, to, types.ConnectivityBlocked)
			}
		}
	}

	return &Result{
		Request: request,
		Table:   resultTable,
	}
}
