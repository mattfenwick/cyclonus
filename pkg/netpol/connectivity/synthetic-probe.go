package connectivity

import (
	"github.com/mattfenwick/cyclonus/pkg/netpol/matcher"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ProtocolPort struct {
	Protocol v1.Protocol
	Port     int
	//Port     intstr.IntOrString // TODO how to support
}

type SyntheticProbeResult struct {
	Port     *ProtocolPort
	Policies *matcher.Policy
	Model    *PodModel
	Ingress  *TruthTable
	Egress   *TruthTable
	Combined *TruthTable
}

func RunSyntheticProbe(policies *matcher.Policy, portProtocol *ProtocolPort, model *PodModel) *SyntheticProbeResult {
	ingressTable := model.NewTruthTable()
	egressTable := model.NewTruthTable()
	combined := model.NewTruthTable()

	log.Infof("running probe on port %d, protocol %s", portProtocol.Port, portProtocol.Protocol)

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
					Protocol: portProtocol.Protocol,
					Port:     intstr.FromInt(portProtocol.Port),
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
		Port:     portProtocol,
		Policies: policies,
		Model:    model,
		Ingress:  ingressTable,
		Egress:   egressTable,
		Combined: combined,
	}
}

func RunSyntheticProbes(policies *matcher.Policy, ports []*ProtocolPort, model *PodModel) []*SyntheticProbeResult {
	var results []*SyntheticProbeResult
	for _, portProtocol := range ports {
		results = append(results, RunSyntheticProbe(policies, portProtocol, model))
	}
	return results
}
