package synthetic

import (
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func RunSyntheticProbe(request *Request) *Result {
	resources := request.Resources
	ingressTable := resources.NewTruthTable()
	egressTable := resources.NewTruthTable()
	combined := resources.NewTruthTable()

	log.Infof("running synthetic probe on port %d, protocol %s", request.Port, request.Protocol)

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
					Port:     intstr.FromInt(request.Port),
				},
			}

			fr := podFrom.PodString().String()
			to := podTo.PodString().String()
			allowed := request.Policies.IsTrafficAllowed(traffic)
			// TODO could also keep the whole `allowed` struct somewhere
			combined.Set(fr, to, allowed.IsAllowed())
			ingressTable.Set(fr, to, allowed.Ingress.IsAllowed)
			egressTable.Set(fr, to, allowed.Egress.IsAllowed)
		}
	}

	return &Result{
		Request:  request,
		Ingress:  ingressTable,
		Egress:   egressTable,
		Combined: combined,
	}
}
