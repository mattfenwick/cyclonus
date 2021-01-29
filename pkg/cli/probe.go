package cli

import (
	"github.com/mattfenwick/cyclonus/pkg/connectivity"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/yaml"
)

type ProbeArgs struct {
	Namespaces                []string
	Pods                      []string
	Noisy                     bool
	IgnoreLoopback            bool
	KubeContext               string
	NetpolCreationWaitSeconds int
	PolicyPath                string
	Port                      int
	Protocol                  string
}

func SetupProbeCommand() *cobra.Command {
	args := &ProbeArgs{}

	command := &cobra.Command{
		Use:   "probe",
		Short: "run a connectivity probe against kubernetes pods",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunProbeCommand(args)
		},
	}

	command.Flags().StringSliceVar(&args.Namespaces, "namespaces", []string{"x", "y", "z"}, "namespaces to create/use pods in")
	command.Flags().StringSliceVar(&args.Pods, "pods", []string{"a", "b", "c"}, "pods to create in namespaces")

	command.Flags().IntVar(&args.Port, "port", 80, "port to run probes on")
	command.Flags().StringVar(&args.Protocol, "protocol", string(v1.ProtocolTCP), "protocol to run probes on")

	command.Flags().BoolVar(&args.Noisy, "noisy", false, "if true, print all results")
	command.Flags().BoolVar(&args.IgnoreLoopback, "ignore-loopback", false, "if true, ignore loopback for truthtable correctness verification")
	command.Flags().StringVar(&args.KubeContext, "kube-context", "", "kubernetes context to use; if empty, uses default context")
	command.Flags().IntVar(&args.NetpolCreationWaitSeconds, "netpol-creation-wait-seconds", 15, "number of seconds to wait after creating a network policy before running probes, to give the CNI time to update the cluster state")
	command.Flags().StringVar(&args.PolicyPath, "policy-path", "", "path to yaml network policy to create in kube; if empty, will not create any policies")

	return command
}

func RunProbeCommand(args *ProbeArgs) {
	if len(args.Namespaces) == 0 || len(args.Pods) == 0 {
		panic(errors.Errorf("found 0 namespaces or pods, must have at least 1 of each"))
	}

	kubernetes, err := kube.NewKubernetes(args.KubeContext)
	utils.DoOrDie(err)

	port := args.Port
	protocol, err := kube.ParseProtocol(args.Protocol)
	utils.DoOrDie(err)

	interpreter, err := connectivity.NewInterpreter(kubernetes, args.Namespaces, args.Pods, port, protocol, false, false)
	utils.DoOrDie(err)

	actions := []*generator.Action{generator.ReadNetworkPolicies(args.Namespaces)}

	if args.PolicyPath != "" {
		policyBytes, err := ioutil.ReadFile(args.PolicyPath)
		utils.DoOrDie(err)

		var kubePolicy networkingv1.NetworkPolicy
		err = yaml.Unmarshal(policyBytes, &kubePolicy)
		utils.DoOrDie(err)

		actions = append(actions, generator.CreatePolicy(&kubePolicy))
	}

	result := interpreter.ExecuteTestCase(generator.NewTestCase(port, protocol, actions))

	printer := connectivity.Printer{
		Noisy:          args.Noisy,
		IgnoreLoopback: args.IgnoreLoopback,
	}
	printer.PrintTestCaseResult(result)
}
