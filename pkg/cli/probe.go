package cli

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/netpol/connectivity"
	"github.com/mattfenwick/cyclonus/pkg/netpol/matcher"
	"github.com/mattfenwick/cyclonus/pkg/netpol/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

type ProbeArgs struct {
	Namespaces     []string
	Pods           []string
	Noisy          bool
	IgnoreLoopback bool
	KubeContext    string
	// TODO
	//Ports                     []int
	//Protocols                 []string
}

func setupProbeCommand() *cobra.Command {
	args := &ProbeArgs{}

	command := &cobra.Command{
		Use:   "probe",
		Short: "run a connectivity probe against kubernetes pods",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			runProbeCommand(args)
		},
	}

	command.Flags().StringSliceVar(&args.Namespaces, "namespaces", []string{"x", "y", "z"}, "namespaces to create/use pods in")
	command.Flags().StringSliceVar(&args.Pods, "pods", []string{"a", "b", "c"}, "pods to create in namespaces")

	// TODO
	//command.Flags().IntSliceVar(&args.Ports, "ports", []int{80}, "ports to run probes on")
	//command.Flags().StringSliceVar(&args.Protocols, "protocols", []string{"tcp"}, "protocols to run probes on")

	command.Flags().BoolVar(&args.Noisy, "noisy", false, "if true, print all results")
	command.Flags().BoolVar(&args.IgnoreLoopback, "ignore-loopback", false, "if true, ignore loopback for truthtable correctness verification")
	command.Flags().StringVar(&args.KubeContext, "kube-context", "", "kubernetes context to use; if empty, uses default context")

	return command
}

func runProbeCommand(args *ProbeArgs) {
	//if len(args.Ports) == 0 || len(args.Protocols) == 0 {
	//	panic(errors.Errorf("found 0 ports or protocols, must have at least 1 of each"))
	//}
	if len(args.Namespaces) == 0 || len(args.Pods) == 0 {
		panic(errors.Errorf("found 0 namespaces or pods, must have at least 1 of each"))
	}

	var kubernetes *kube.Kubernetes
	var err error
	if args.KubeContext == "" {
		kubernetes, err = kube.NewKubernetesForDefaultContext()
	} else {
		kubernetes, err = kube.NewKubernetesForContext(args.KubeContext)
	}
	utils.DoOrDie(err)

	port := &connectivity.ProtocolPort{
		Protocol: v1.ProtocolTCP,
		Port:     80,
	}
	podModel := connectivity.NewDefaultModel(args.Namespaces, args.Pods, port.Port, port.Protocol)

	utils.DoOrDie(connectivity.CreateResources(kubernetes, podModel))
	waitForPodsReady(kubernetes, args.Namespaces, args.Pods, 60)

	podList, err := kubernetes.GetPodsInNamespaces(args.Namespaces)
	utils.DoOrDie(err)
	for _, pod := range podList {
		ip := pod.Status.PodIP
		if ip == "" {
			panic(errors.Errorf("no ip found for pod %s/%s", pod.Namespace, pod.Name))
		}
		podModel.Namespaces[pod.Namespace].Pods[pod.Name].IP = ip
		log.Infof("ip for pod %s/%s: %s", pod.Namespace, pod.Name, ip)
	}

	// read policies from kube
	kubePolicies, err := kubernetes.GetNetworkPoliciesInNamespaces(args.Namespaces)
	utils.DoOrDie(err)
	kubePoliciesPointers := make([]*networkingv1.NetworkPolicy, len(kubePolicies))
	for i := range kubePolicies {
		kubePoliciesPointers[i] = &kubePolicies[i]
	}
	log.Infof("found %d policies across namespaces %+v", len(kubePolicies), args.Namespaces)
	policy := matcher.BuildNetworkPolicies(kubePoliciesPointers)

	if args.Noisy {
		fmt.Printf("%s\n\n", matcher.Explain(policy))
		matcher.TableExplainer(policy).Render()
	}

	log.Infof("synthetic probe on port %d, protocol %s", port.Port, port.Protocol)
	synthetic := connectivity.RunSyntheticProbe(policy, port, podModel)

	log.Infof("kube probe on port %d, protocol %s", port.Port, port.Protocol)
	kubeProbe := connectivity.RunKubeProbe(kubernetes, podModel, port.Port, port.Protocol, 5)

	fmt.Printf("\n\nKube results:\n")
	kubeProbe.Table().Render()

	comparison := synthetic.Combined.Compare(kubeProbe)
	t, f, nv, checked := comparison.ValueCounts(args.IgnoreLoopback)
	if f > 0 {
		fmt.Printf("Discrepancy found: %d wrong, %d no value, %d correct out of %d total\n", f, t, nv, checked)
	} else {
		fmt.Printf("found %d true, %d false, %d no value from %d total\n", t, f, nv, checked)
	}

	if f > 0 || args.Noisy {
		fmt.Println("Ingress:")
		synthetic.Ingress.Table().Render()

		fmt.Println("Egress:")
		synthetic.Egress.Table().Render()

		fmt.Println("Combined:")
		synthetic.Combined.Table().Render()

		fmt.Printf("\n\nSynthetic vs combined:\n")
		comparison.Table().Render()
	}
}
