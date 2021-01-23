package cli

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/connectivity"
	kube2 "github.com/mattfenwick/cyclonus/pkg/connectivity/kube"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

type GeneratorArgs struct {
	Mode                      string
	AllowDNS                  bool
	Noisy                     bool
	IgnoreLoopback            bool
	NetpolCreationWaitSeconds int
	KubeContext               string
}

func SetupGeneratorCommand() *cobra.Command {
	args := &GeneratorArgs{}

	command := &cobra.Command{
		Use:   "generator",
		Short: "generate network policies",
		Long:  "generate network policies including corner cases by combinations of fragments",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunGeneratorCommand(args)
		},
	}

	command.Flags().StringVar(&args.Mode, "mode", "", "mode used to generate network policies")
	utils.DoOrDie(command.MarkFlagRequired("mode"))

	command.Flags().BoolVar(&args.AllowDNS, "allow-dns", true, "if using egress, allow udp over port 53 for DNS resolution")
	command.Flags().BoolVar(&args.Noisy, "noisy", false, "if true, print all results")
	command.Flags().BoolVar(&args.IgnoreLoopback, "ignore-loopback", false, "if true, ignore loopback for truthtable correctness verification")
	command.Flags().IntVar(&args.NetpolCreationWaitSeconds, "netpol-creation-wait-seconds", 5, "number of seconds to wait after creating a network policy before running probes, to give the CNI time to update the cluster state")
	command.Flags().StringVar(&args.KubeContext, "kube-context", "", "kubernetes context to use; if empty, uses default context")

	return command
}

func RunGeneratorCommand(args *GeneratorArgs) {
	namespaces := []string{"x", "y", "z"}
	pods := []string{"a", "b", "c"}

	port := &connectivity.ProtocolPort{
		Protocol: v1.ProtocolTCP,
		Port:     80,
	}
	// TODO don't use default model?
	podModel := connectivity.NewDefaultModel(namespaces, pods, port.Port, port.Protocol)
	var kubernetes *kube.Kubernetes
	var err error
	if args.KubeContext == "" {
		kubernetes, err = kube.NewKubernetesForDefaultContext()
	} else {
		kubernetes, err = kube.NewKubernetesForContext(args.KubeContext)
	}
	utils.DoOrDie(err)

	utils.DoOrDie(kube2.CreateResources(kubernetes, podModel))
	waitForPodsReady(kubernetes, namespaces, pods, 60)

	podList, err := kubernetes.GetPodsInNamespaces(namespaces)
	utils.DoOrDie(err)
	for _, pod := range podList {
		ip := pod.Status.PodIP
		if ip == "" {
			panic(errors.Errorf("no ip found for pod %s/%s", pod.Namespace, pod.Name))
		}
		podModel.Namespaces[pod.Namespace].Pods[pod.Name].IP = ip
		log.Infof("ip for pod %s/%s: %s", pod.Namespace, pod.Name, ip)
	}

	zcPod, err := kubernetes.GetPod("z", "c")
	utils.DoOrDie(err)
	if zcPod.Status.PodIP == "" {
		panic(errors.Errorf("no ip found for pod z/c"))
	}
	zcIP := zcPod.Status.PodIP

	generator := &generator.FragmentGenerator{
		Ports:            generator.DefaultPorts(),
		PodPeers:         generator.DefaultPodPeers(zcIP),
		Targets:          generator.DefaultTargets(),
		Namespaces:       namespaces,
		TypicalPorts:     generator.TypicalPorts,
		TypicalPeers:     generator.TypicalPeers,
		TypicalTarget:    generator.TypicalTarget,
		TypicalNamespace: generator.TypicalNamespace,
	}

	var kubePolicies []*networkingv1.NetworkPolicy
	switch args.Mode {
	//case "ingress-egress":
	//	kubePolicies = generator.IngressEgressPolicies(args.AllowDNS)
	case "ingress":
		kubePolicies = generator.IngressPolicies()
	case "egress":
		kubePolicies = generator.EgressPolicies(args.AllowDNS)
	case "vary-ingress": // TODO come up with a better name
		kubePolicies = generator.VaryIngressPolicies()
	case "vary-egress": // TODO come up with a better name
		kubePolicies = generator.VaryEgressPolicies(args.AllowDNS)
	default:
		panic(errors.Errorf("invalid test mode %s", args.Mode))
	}
	fmt.Printf("testing %d policies\n\n", len(kubePolicies))

	namespacesToCleanSet := map[string]bool{}
	namespacesToClean := []string{}
	for _, kp := range kubePolicies {
		if !namespacesToCleanSet[kp.Namespace] {
			namespacesToCleanSet[kp.Namespace] = true
			namespacesToClean = append(namespacesToClean, kp.Namespace)
		}
	}
	for _, ns := range namespaces {
		if !namespacesToCleanSet[ns] {
			namespacesToCleanSet[ns] = true
			namespacesToClean = append(namespacesToClean, ns)
		}
	}

	tester := connectivity.NewTester(kubernetes)

	for i, kubePolicy := range kubePolicies {
		testCase := &connectivity.TestCase{
			KubePolicy:                kubePolicy,
			Noisy:                     args.Noisy,
			NetpolCreationWaitSeconds: args.NetpolCreationWaitSeconds,
			Port:                      port.Port,
			Protocol:                  port.Protocol,
			PodModel:                  podModel,
			IgnoreLoopback:            args.IgnoreLoopback,
			NamespacesToClean:         namespacesToClean,
		}
		result := tester.TestNetworkPolicy(testCase)
		utils.DoOrDie(result.Err)

		fmt.Printf("\nfinished policy #%d\n\n", i)
	}
}
