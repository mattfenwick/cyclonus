package cli

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/connectivity"
	connectivitykube "github.com/mattfenwick/cyclonus/pkg/connectivity/kube"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/synthetic"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"sort"
)

type CompareArgs struct {
	Noisy                     bool
	NetpolCreationWaitSeconds int
	Contexts                  []string
}

func SetupCompareCommand() *cobra.Command {
	args := &CompareArgs{}

	command := &cobra.Command{
		Use:   "compare",
		Short: "compare network policy",
		Long:  "Compare network policies between multiple clusters",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunCompareCommand(args)
		},
	}

	command.Flags().BoolVar(&args.Noisy, "noisy", false, "if true, print all results")
	command.Flags().IntVar(&args.NetpolCreationWaitSeconds, "netpol-creation-wait-seconds", 5, "number of seconds to wait after creating a network policy before running probes, to give the CNI time to update the cluster state")
	command.Flags().StringSliceVar(&args.Contexts, "context", []string{}, "kubernetes context to use; if empty, uses default context")

	return command
}

func RunCompareCommand(args *CompareArgs) {
	namespaces := []string{"x", "y", "z"}
	pods := []string{"a", "b", "c"}

	protocol := v1.ProtocolTCP
	port := 80

	kubeClients := map[string]*kube.Kubernetes{}
	if len(args.Contexts) == 0 {
		kubernetes, err := kube.NewKubernetesForContext("")
		utils.DoOrDie(err)
		kubeClients["default-context"] = kubernetes
	} else {
		for _, context := range args.Contexts {
			kubernetes, err := kube.NewKubernetesForContext(context)
			utils.DoOrDie(err)
			kubeClients[context] = kubernetes
		}
	}

	kubeResources := map[string]*connectivitykube.Resources{}
	var syntheticResources *synthetic.Resources
	var zcIP string
	for context, kubeClient := range kubeClients {
		kubernetesResources, synth, err := connectivity.SetupClusterTODODelete(kubeClient, namespaces, pods, port, protocol)
		utils.DoOrDie(err)
		// TODO this is a huge hack -- ips are going to be different from cluster to cluster, which means
		//   that policies involving ips need to be different from cluster to cluster.  But here we're just
		//   taking the first one and using it everywhere.
		if syntheticResources == nil {
			syntheticResources = synth

			zcPod, err := kubeClient.GetPod("z", "c")
			utils.DoOrDie(err)
			if zcPod.Status.PodIP == "" {
				panic(errors.Errorf("no ip found for pod z/c"))
			}
			zcIP = zcPod.Status.PodIP
		}

		kubeResources[context] = kubernetesResources
	}

	fragGenerator := generator.NewDefaultFragmentGenerator(true, namespaces, zcIP)
	kubePolicySlices := packIntoSlices(fragGenerator.FragmentPolicies())

	fmt.Printf("testing %d policies\n\n", len(kubePolicySlices))

	contexts := args.Contexts
	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i] < contexts[j]
	})
	tester := connectivity.NewMultipleContextTester()
	printer := &connectivity.MultipleContextTestCasePrinter{
		Noisy:    args.Noisy,
		Contexts: contexts,
	}

	for i, kubePolicy := range kubePolicySlices {
		logrus.Infof("starting policy #%d", i)
		testCase := &connectivity.MultipleContextTestCase{
			KubePolicies:              kubePolicy,
			NetpolCreationWaitSeconds: args.NetpolCreationWaitSeconds,
			Port:                      port,
			Protocol:                  protocol,
			KubeClients:               kubeClients,
			KubeResources:             kubeResources,
			SyntheticResources:        syntheticResources,
			NamespacesToClean:         namespaces,
			Policy:                    matcher.BuildNetworkPolicies(kubePolicy),
		}
		result := tester.TestNetworkPolicy(testCase)
		if len(result.Errors) > 0 {
			panic(result.Errors[0])
		}

		printer.PrintTestCaseResult(result)
		logrus.Infof("finished policy #%d", i)
	}

	printer.PrintFinish()
}

func packIntoSlices(netpols []*networkingv1.NetworkPolicy) [][]*networkingv1.NetworkPolicy {
	var sos [][]*networkingv1.NetworkPolicy
	for _, np := range netpols {
		sos = append(sos, []*networkingv1.NetworkPolicy{np})
	}
	return sos
}
