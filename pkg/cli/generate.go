package cli

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/connectivity"
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
	NetpolCreationWaitSeconds int
	Context                   string
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

	command.Flags().BoolVar(&args.AllowDNS, "allow-dns", true, "if using egress, allow udp over port 53 for DNS resolution")
	command.Flags().BoolVar(&args.Noisy, "noisy", false, "if true, print all results")
	command.Flags().BoolVar(&args.IgnoreLoopback, "ignore-loopback", false, "if true, ignore loopback for truthtable correctness verification")
	command.Flags().IntVar(&args.NetpolCreationWaitSeconds, "netpol-creation-wait-seconds", 5, "number of seconds to wait after creating a network policy before running probes, to give the CNI time to update the cluster state")
	command.Flags().StringVar(&args.Context, "context", "", "kubernetes context to use; if empty, uses default context")

	return command
}

func RunGenerateCommand(args *GenerateArgs) {
	namespaces := []string{"x", "y", "z"}
	pods := []string{"a", "b", "c"}

	protocol := v1.ProtocolTCP
	port := 80

	kubernetes, err := kube.NewKubernetes(args.Context)
	utils.DoOrDie(err)

	zcPod, err := kubernetes.GetPod("z", "c")
	utils.DoOrDie(err)
	if zcPod.Status.PodIP == "" {
		panic(errors.Errorf("no ip found for pod z/c"))
	}
	zcIP := zcPod.Status.PodIP

	var testCaseGenerator generator.TestCaseGenerator
	switch args.Mode {
	case "simple-fragments":
		testCaseGenerator = generator.NewDefaultFragmentGenerator(args.AllowDNS, namespaces, zcIP)
	case "conflicts":
		testCaseGenerator = &generator.ConflictGenerator{
			AllowDNS:    args.AllowDNS,
			Source:      generator.NewNetpolTarget("x", map[string]string{"pod": "b"}, nil),
			Destination: generator.NewNetpolTarget("y", map[string]string{"pod": "c"}, nil)}
	default:
		panic(errors.Errorf("invalid test mode %s", args.Mode))
	}

	interpreter, err := connectivity.NewInterpreter(kubernetes, namespaces, pods, port, protocol, true, false, true)
	utils.DoOrDie(err)
	printer := &connectivity.Printer{
		Noisy:          args.Noisy,
		IgnoreLoopback: args.IgnoreLoopback,
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
}
