package cli

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/connectivity"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/probe"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type GenerateArgs struct {
	Mode                      string
	AllowDNS                  bool
	Noisy                     bool
	IgnoreLoopback            bool
	PerturbationWaitSeconds   int
	PodCreationTimeoutSeconds int
	Context                   string
	ServerPorts               []int
	ServerProtocols           []string
	ServerNamespaces          []string
	ServerPods                []string
	CleanupNamespaces         bool
}

func SetupGenerateCommand() *cobra.Command {
	args := &GenerateArgs{}

	command := &cobra.Command{
		Use:   "generate",
		Short: "generate network policies",
		Long:  "generate network policies, create and probe against kubernetes, and compare to expected results",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunGenerateCommand(args)
		},
	}

	command.Flags().StringVar(&args.Mode, "mode", "", "mode used to generate network policies")
	utils.DoOrDie(command.MarkFlagRequired("mode"))

	// TODO add UDP to defaults once support has been added
	command.Flags().StringSliceVar(&args.ServerProtocols, "server-protocol", []string{"tcp", "sctp"}, "protocols to run server on")
	command.Flags().IntSliceVar(&args.ServerPorts, "server-port", []int{80, 81}, "ports to run server on")
	command.Flags().StringSliceVar(&args.ServerNamespaces, "namespace", []string{"x", "y", "z"}, "namespaces to create/use pods in")
	command.Flags().StringSliceVar(&args.ServerPods, "pod", []string{"a", "b", "c"}, "pods to create in namespaces")

	command.Flags().BoolVar(&args.AllowDNS, "allow-dns", true, "if using egress, allow udp over port 53 for DNS resolution")
	command.Flags().BoolVar(&args.Noisy, "noisy", false, "if true, print all results")
	command.Flags().BoolVar(&args.IgnoreLoopback, "ignore-loopback", false, "if true, ignore loopback for truthtable correctness verification")
	command.Flags().IntVar(&args.PerturbationWaitSeconds, "perturbation-wait-seconds", 5, "number of seconds to wait after perturbing the cluster (i.e. create a network policy, modify a ns/pod label) before running probes, to give the CNI time to update the cluster state")
	command.Flags().IntVar(&args.PodCreationTimeoutSeconds, "pod-creation-timeout-seconds", 60, "number of seconds to wait for pods to create, be running and have IP addresses")
	command.Flags().StringVar(&args.Context, "context", "", "kubernetes context to use; if empty, uses default context")
	command.Flags().BoolVar(&args.CleanupNamespaces, "cleanup-namespaces", false, "if true, clean up namespaces after completion")

	return command
}

func RunGenerateCommand(args *GenerateArgs) {
	externalIPs := []string{} // "http://www.google.com"} // TODO make these be IPs?  or not?

	kubernetes, err := kube.NewKubernetesForContext(args.Context)
	utils.DoOrDie(err)

	serverProtocols := parseProtocols(args.ServerProtocols)

	resources, err := probe.NewDefaultResources(kubernetes, args.ServerNamespaces, args.ServerPods, args.ServerPorts, serverProtocols, externalIPs, args.PodCreationTimeoutSeconds)
	utils.DoOrDie(err)
	interpreter, err := connectivity.NewInterpreter(kubernetes, resources, true, 1, args.PerturbationWaitSeconds, true)
	utils.DoOrDie(err)
	printer := &connectivity.Printer{
		Noisy:          args.Noisy,
		IgnoreLoopback: args.IgnoreLoopback,
	}

	zcPod, err := resources.GetPod("z", "c")
	utils.DoOrDie(err)

	var testCaseGenerator generator.TestCaseGenerator
	switch args.Mode {
	case "example":
		testCaseGenerator = &generator.ExampleGenerator{}
	case "upstream":
		testCaseGenerator = &generator.UpstreamE2EGenerator{}
	case "simple-fragments":
		testCaseGenerator = generator.NewDefaultFragmentGenerator(args.AllowDNS, args.ServerNamespaces, zcPod.IP)
	case "discrete":
		testCaseGenerator = generator.NewDefaultDiscreteGenerator(args.AllowDNS, zcPod.IP)
	case "conflicts":
		testCaseGenerator = &generator.ConflictGenerator{
			AllowDNS:    args.AllowDNS,
			Source:      generator.NewNetpolTarget("x", map[string]string{"pod": "b"}, nil),
			Destination: generator.NewNetpolTarget("y", map[string]string{"pod": "c"}, nil)}
	default:
		panic(errors.Errorf("invalid test mode %s", args.Mode))
	}

	testCases := testCaseGenerator.GenerateTestCases()
	fmt.Printf("testing %d cases\n\n", len(testCases))
	for i, testCase := range testCases {
		logrus.Infof("starting test case #%d", i+1)

		result := interpreter.ExecuteTestCase(testCase)
		utils.DoOrDie(result.Err)

		printer.PrintTestCaseResult(result)
		logrus.Infof("finished policy #%d", i+1)
	}

	printer.PrintSummary()

	if args.CleanupNamespaces {
		for _, ns := range args.ServerNamespaces {
			logrus.Infof("cleaning up namespace %s", ns)
			err = kubernetes.DeleteNamespace(ns)
			if err != nil {
				logrus.Warnf("%+v", err)
			}
		}
	}
}
