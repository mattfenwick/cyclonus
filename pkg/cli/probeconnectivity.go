package cli

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/netpol/connectivity"
	"github.com/mattfenwick/cyclonus/pkg/netpol/matcher"
	"github.com/mattfenwick/cyclonus/pkg/netpol/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ProbeConnectivityArgs struct {
	PolicySource    string
	Namespaces      []string
	PolicyPath      string
	ModelNamespaces []string
	ModelPods       []string
	Protocol        string
	NumberedPort    int
	NamedPort       string
}

func setupProbeConnectivityCommand() *cobra.Command {
	args := &ProbeConnectivityArgs{}

	command := &cobra.Command{
		Use:   "probe",
		Short: "probe connectivity",
		Long:  "probe connectivity against a cluster model; does not use a real cluster",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			runProbeConnectivityCommand(args)
		},
	}

	command.Flags().StringVar(&args.PolicySource, "policy-source", "kube", "source of network policies (kube, file, examples)")

	command.Flags().StringSliceVar(&args.Namespaces, "namespaces", []string{}, "only set if policy-source = kube; selects namespaces to read policies from; leaving empty will select all namespaces")

	command.Flags().StringVar(&args.PolicyPath, "policy-path", "", "only set if policy-source = file; path to network polic(ies)")

	command.Flags().StringSliceVar(&args.ModelNamespaces, "model-namespace", []string{"x", "y", "z"}, "namespaces to use in model")
	command.Flags().StringSliceVar(&args.ModelPods, "model-pods", []string{"a", "b", "c"}, "pods to use in model")

	command.Flags().StringVar(&args.Protocol, "protocol", string(v1.ProtocolTCP), "protocol to run probe over")
	command.Flags().IntVar(&args.NumberedPort, "numbered-port", 0, "numbered port to run probe over")
	command.Flags().StringVar(&args.NamedPort, "named-port", "", "named port to run probe over")

	return command
}

func runProbeConnectivityCommand(args *ProbeConnectivityArgs) {
	// 1. source of policies
	kubePolicies, err := readPolicies(args.PolicySource, args.Namespaces, args.PolicyPath)
	utils.DoOrDie(err)

	// 2. consume policies
	explainedPolicies := matcher.BuildNetworkPolicies(kubePolicies)

	// 3. use model to create traffic
	model := connectivity.NewModel(args.ModelNamespaces, args.ModelPods)
	ingressTable := model.NewTruthTable()
	egressTable := model.NewTruthTable()
	truthTable := model.NewTruthTable()

	var protocol v1.Protocol
	switch args.Protocol {
	case "udp", "UDP":
		protocol = v1.ProtocolUDP
	case "sctp", "SCTP":
		protocol = v1.ProtocolSCTP
	case "tcp", "TCP":
		protocol = v1.ProtocolTCP
	}

	var port intstr.IntOrString
	if args.NumberedPort != 0 && args.NamedPort != "" {
		panic(errors.Errorf("both numbered-port and named-port were set -- but only one may be set"))
	} else if args.NumberedPort != 0 {
		port = intstr.FromInt(args.NumberedPort)
	} else if args.NamedPort != "" {
		port = intstr.FromString(args.NamedPort)
	} else {
		panic(errors.Errorf("neither numbered-port nor named-port were set -- but one must be set"))
	}

	// TODO add ips
	for _, namespaceFrom := range model.Namespaces {
		for _, podFrom := range namespaceFrom.Pods {
			for _, namespaceTo := range model.Namespaces {
				for _, podTo := range namespaceTo.Pods {
					traffic := &matcher.Traffic{
						Source: &matcher.TrafficPeer{
							Internal: &matcher.InternalPeer{
								PodLabels:       podFrom.Labels,
								NamespaceLabels: namespaceFrom.Labels,
								Namespace:       namespaceFrom.Name,
							},
							//IP:       "", TODO
						},
						Destination: &matcher.TrafficPeer{
							Internal: &matcher.InternalPeer{
								PodLabels:       podTo.Labels,
								NamespaceLabels: namespaceTo.Labels,
								Namespace:       namespaceTo.Name,
							},
							//IP:       "", TODO
						},
						PortProtocol: &matcher.PortProtocol{
							Protocol: protocol,
							Port:     port,
						},
					}
					fr := podFrom.PodString().String()
					to := podTo.PodString().String()
					allowed := explainedPolicies.IsTrafficAllowed(traffic)
					truthTable.Set(fr, to, allowed.IsAllowed())
					ingressTable.Set(fr, to, allowed.Ingress.IsAllowed)
					egressTable.Set(fr, to, allowed.Egress.IsAllowed)
				}
			}
		}
	}

	// 4. print out a result matrix
	fmt.Println("Ingress:")
	ingressTable.Table().Render()
	fmt.Println("Egress:")
	egressTable.Table().Render()
	fmt.Println("Combined:")
	truthTable.Table().Render()
}
