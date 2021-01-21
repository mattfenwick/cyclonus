package connectivity

import (
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type SyntheticProbeResult struct {
	Protocol v1.Protocol
	Port     int
	Policies *matcher.Policy
	Model    *PodModel
	Ingress  *TruthTable
	Egress   *TruthTable
	Combined *TruthTable
}

func RunSyntheticProbe(policies *matcher.Policy, protocol v1.Protocol, port int, model *PodModel) *SyntheticProbeResult {
	ingressTable := model.NewTruthTable()
	egressTable := model.NewTruthTable()
	combined := model.NewTruthTable()

	log.Infof("running probe on port %d, protocol %s", port, protocol)

	for _, podFrom := range model.AllPods() {
		for _, podTo := range model.AllPods() {
			traffic := &matcher.Traffic{
				Source: &matcher.TrafficPeer{
					Internal: &matcher.InternalPeer{
						PodLabels:       podFrom.Pod.Labels,
						NamespaceLabels: podFrom.Namespace.Labels,
						Namespace:       podFrom.NamespaceName,
					},
					IP: podFrom.Pod.IP,
				},
				Destination: &matcher.TrafficPeer{
					Internal: &matcher.InternalPeer{
						PodLabels:       podTo.Pod.Labels,
						NamespaceLabels: podTo.Namespace.Labels,
						Namespace:       podTo.NamespaceName,
					},
					IP: podTo.Pod.IP,
				},
				PortProtocol: &matcher.PortProtocol{
					Protocol: protocol,
					Port:     intstr.FromInt(port),
				},
			}

			fr := podFrom.PodString().String()
			to := podTo.PodString().String()
			allowed := policies.IsTrafficAllowed(traffic)
			combined.Set(fr, to, allowed.IsAllowed())
			ingressTable.Set(fr, to, allowed.Ingress.IsAllowed)
			egressTable.Set(fr, to, allowed.Egress.IsAllowed)
		}
	}

	return &SyntheticProbeResult{
		Port:     port,
		Protocol: protocol,
		Policies: policies,
		Model:    model,
		Ingress:  ingressTable,
		Egress:   egressTable,
		Combined: combined,
	}
}

type ProtocolPort struct {
	Port     int
	Protocol v1.Protocol
}

func RunSyntheticProbes(policies *matcher.Policy, ports []*ProtocolPort, model *PodModel) []*SyntheticProbeResult {
	var results []*SyntheticProbeResult
	for _, portProtocol := range ports {
		results = append(results, RunSyntheticProbe(policies, portProtocol.Protocol, portProtocol.Port, model))
	}
	return results
}
