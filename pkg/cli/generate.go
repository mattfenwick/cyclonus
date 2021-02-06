package cli

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/connectivity"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/types"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
)

type GenerateArgs struct {
	Mode                      string
	AllowDNS                  bool
	Noisy                     bool
	IgnoreLoopback            bool
	PerturbationWaitSeconds   int
	PodCreationTimeoutSeconds int
	Context                   string
	Namespaces                []string
	Pods                      []string
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

	command.Flags().StringSliceVar(&args.Namespaces, "namespaces", []string{"x", "y", "z"}, "namespaces to create/use pods in")
	command.Flags().StringSliceVar(&args.Pods, "pods", []string{"a", "b", "c"}, "pods to create in namespaces")

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
	serverProtocols := []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}
	serverPorts := []int{80, 81}
	externalIPs := []string{} // "http://www.google.com"} // TODO make these be IPs?  or not?

	kubernetes, err := kube.NewKubernetesForContext(args.Context)
	utils.DoOrDie(err)

	kubeResources := types.NewDefaultResources(args.Namespaces, args.Pods, serverPorts, serverProtocols, externalIPs)
	interpreter, err := connectivity.NewInterpreter(kubernetes, kubeResources, true, 1, args.PerturbationWaitSeconds, args.PodCreationTimeoutSeconds, true)
	utils.DoOrDie(err)
	printer := &connectivity.Printer{
		Noisy:          args.Noisy,
		IgnoreLoopback: args.IgnoreLoopback,
	}

	zcPod, err := kubernetes.GetPod("z", "c")
	utils.DoOrDie(err)
	if zcPod.Status.PodIP == "" {
		panic(errors.Errorf("no ip found for pod z/c"))
	}
	zcIP := zcPod.Status.PodIP

	var testCaseGenerator generator.TestCaseGenerator
	switch args.Mode {
	case "upstream":
		testCaseGenerator = &generator.UpstreamE2EGenerator{}
	case "simple-fragments":
		testCaseGenerator = generator.NewDefaultFragmentGenerator(args.AllowDNS, args.Namespaces, zcIP)
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
		logrus.Infof("starting test case #%d", i)

		result := interpreter.ExecuteTestCase(testCase)
		utils.DoOrDie(result.Err)

		printer.PrintTestCaseResult(result)
		logrus.Infof("finished policy #%d", i)
	}

	printer.PrintSummary()

	if args.CleanupNamespaces {
		for _, ns := range args.Namespaces {
			logrus.Infof("cleaning up namespace %s", ns)
			err = kubernetes.DeleteNamespace(ns)
			if err != nil {
				logrus.Warnf("%+v", err)
			}
		}
	}
}
