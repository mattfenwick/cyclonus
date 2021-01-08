package cli

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/netpol/connectivity"
	"github.com/mattfenwick/cyclonus/pkg/netpol/matcher"
	"github.com/mattfenwick/cyclonus/pkg/netpol/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"os"
)

func RunRootCommand() {
	command := setupRootCommand()
	if err := errors.Wrapf(command.Execute(), "run root command"); err != nil {
		log.Fatalf("unable to run root command: %+v", err)
		os.Exit(1)
	}
}

type Flags struct {
	Verbosity string
}

func setupRootCommand() *cobra.Command {
	flags := &Flags{}
	command := &cobra.Command{
		Use:   "netpol-explainer",
		Short: "explain, analyze, and query network policies",
		Long:  "explain, analyze, and query network policies",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return SetUpLogger(flags.Verbosity)
		},
	}

	command.PersistentFlags().StringVarP(&flags.Verbosity, "verbosity", "v", "info", "log level; one of [info, debug, trace, warn, error, fatal, panic]")

	command.AddCommand(setupAnalyzePoliciesCommand())
	command.AddCommand(setupQueryTrafficCommand())
	command.AddCommand(setupProbeConnectivityCommand())

	// TODO
	//command.AddCommand(setupQueryTargetsCommand())
	//command.AddCommand(setupQueryPeersCommand())

	return command
}

type AnalyzePoliciesArgs struct {
	PolicySource string
	Namespaces   []string
	PolicyPath   string
	Format       string
}

func setupAnalyzePoliciesCommand() *cobra.Command {
	args := &AnalyzePoliciesArgs{}

	command := &cobra.Command{
		Use:   "analyze",
		Short: "analyze network policies",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			runAnalyzePoliciesCommand(args)
		},
	}

	command.Flags().StringVar(&args.PolicySource, "policy-source", "kube", "source of network policies (kube, file, examples)")

	command.Flags().StringSliceVar(&args.Namespaces, "namespaces", []string{}, "only set if policy-source = kube; selects namespaces to read policies from; leaving empty will select all namespaces")

	command.Flags().StringVar(&args.PolicyPath, "policy-path", "", "only set if policy-source = file; path to network polic(ies)")

	command.Flags().StringVar(&args.Format, "format", "", "output format; human-readable if empty (options: json)")

	return command
}

func runAnalyzePoliciesCommand(args *AnalyzePoliciesArgs) {
	// 1. source of policies
	kubePolicies, err := readPolicies(args.PolicySource, args.Namespaces, args.PolicyPath)
	utils.DoOrDie(err)

	// 2. consume policies
	explainedPolicies := matcher.BuildNetworkPolicies(kubePolicies)
	switch args.Format {
	case "json":
		printJSON(explainedPolicies)
	default:
		fmt.Printf("%s\n\n", matcher.Explain(explainedPolicies))
	}
}

type QueryTrafficArgs struct {
	PolicySource string
	Namespaces   []string
	TrafficFile  string
	PolicyPath   string
}

func setupQueryTrafficCommand() *cobra.Command {
	args := &QueryTrafficArgs{}

	command := &cobra.Command{
		Use:   "traffic",
		Short: "query traffic allow/deny",
		Long:  "given policies and traffic as input, determine whether the traffic would be allowed",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			runQueryTrafficCommand(args)
		},
	}

	command.Flags().StringVar(&args.PolicySource, "policy-source", "kube", "source of network policies (kube, file, examples)")

	command.Flags().StringSliceVar(&args.Namespaces, "namespaces", []string{}, "only set if policy-source = kube; selects namespaces to read policies from; leaving empty will select all namespaces")

	command.Flags().StringVar(&args.PolicyPath, "policy-path", "", "only set if policy-source = file; path to network polic(ies)")

	command.Flags().StringVar(&args.TrafficFile, "traffic-file", "", "path to traffic file, containing a list of traffic objects")
	command.MarkFlagRequired("traffic-file")

	return command
}

func runQueryTrafficCommand(args *QueryTrafficArgs) {
	// 1. source of policies
	kubePolicies, err := readPolicies(args.PolicySource, args.Namespaces, args.PolicyPath)
	utils.DoOrDie(err)

	// 2. consume policies
	explainedPolicies := matcher.BuildNetworkPolicies(kubePolicies)

	// 3. query
	var allTraffics []*matcher.Traffic
	allTrafficBytes, err := ioutil.ReadFile(args.TrafficFile)
	utils.DoOrDie(err)
	err = json.Unmarshal(allTrafficBytes, &allTraffics)
	utils.DoOrDie(err)
	for _, traffic := range allTraffics {
		trafficBytes, err := json.MarshalIndent(traffic, "", "  ")
		utils.DoOrDie(err)
		result := explainedPolicies.IsTrafficAllowed(traffic)
		resultBytes, err := json.MarshalIndent(result, "", "  ")
		utils.DoOrDie(err)
		fmt.Printf("Traffic:\n%s\n\nIs allowed: %t\n\nExplanation:\n\n%s\n\n", string(trafficBytes), result.IsAllowed(), string(resultBytes))
	}
}

type ProbeConnectivityArgs struct {
	PolicySource    string
	Namespaces      []string
	PolicyPath      string
	ModelNamespaces []string
	ModelPods       []string
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

	// TODO add protocol
	// TODO add port
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
							Protocol: v1.ProtocolTCP,
							Port:     intstr.FromInt(80),
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
