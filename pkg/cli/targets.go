package cli

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/netpol/matcher"
	"github.com/mattfenwick/cyclonus/pkg/netpol/utils"
	"github.com/spf13/cobra"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

type QueryTargetsArgs struct {
	PolicySource string
	Namespaces   []string
	PolicyPath   string
	PodPath      string
}

func SetupQueryTargetsCommand() *cobra.Command {
	args := &QueryTargetsArgs{}

	command := &cobra.Command{
		Use:   "targets",
		Short: "query targets",
		Long:  "given a pod with labels in a namespace with labels, determine which targets apply to the pod",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunQueryTargetsCommand(args)
		},
	}

	command.Flags().StringVar(&args.PolicySource, "policy-source", "kube", "source of network policies (kube, file, examples)")

	command.Flags().StringSliceVar(&args.Namespaces, "namespaces", []string{}, "only set if policy-source = kube; selects namespaces to read policies from; leaving empty will select all namespaces")

	command.Flags().StringVar(&args.PolicyPath, "policy-path", "", "only set if policy-source = file; path to network polic(ies)")

	command.Flags().StringVar(&args.PodPath, "pod-path", "", "path to pod file -- json array of dicts")
	utils.DoOrDie(command.MarkFlagRequired("pod-path"))

	return command
}

type QueryTargetPod struct {
	Namespace string
	Labels    map[string]string
}

func RunQueryTargetsCommand(args *QueryTargetsArgs) {
	// 1. read policies
	kubePolicies, err := readPolicies(args.PolicySource, args.Namespaces, args.PolicyPath)
	utils.DoOrDie(err)

	// 2. consume policies
	explainedPolicies := matcher.BuildNetworkPolicies(kubePolicies)

	// 3. read pods
	var pods []QueryTargetPod
	bs, err := ioutil.ReadFile(args.PodPath)
	utils.DoOrDie(err)
	err = json.Unmarshal(bs, &pods)
	utils.DoOrDie(err)

	// 4. query
	for _, pod := range pods {
		fmt.Printf("pod %+v:\n\n", pod)

		// ingress
		fmt.Println("  ingress")
		ingressValue := true
		ingressTargets := explainedPolicies.TargetsApplyingToPod(ingressValue, pod.Namespace, pod.Labels)
		for _, t := range ingressTargets {
			fmt.Printf("    %s\n", strings.Join(matcher.ExplainTarget(t, ingressValue), "\n"))
		}
		// combine all the ingress targets for combined connectivity
		combinedIngressTarget := matcher.CombineTargetsIgnoringPrimaryKey(pod.Namespace, metav1.LabelSelector{MatchLabels: pod.Labels}, ingressTargets)
		if combinedIngressTarget != nil {
			fmt.Printf("    combined ingress:\n%s\n\n", strings.Join(matcher.ExplainTarget(combinedIngressTarget, ingressValue), "\n"))
		} else {
			fmt.Println("    combined ingress: none")
		}

		// egress
		fmt.Printf("\n  egress\n")
		egressValue := false
		egressTargets := explainedPolicies.TargetsApplyingToPod(egressValue, pod.Namespace, pod.Labels)
		for _, t := range egressTargets {
			fmt.Printf("    %s\n", strings.Join(matcher.ExplainTarget(t, egressValue), "\n"))
		}
		// combine all the egress targets for combined connectivity
		combinedEgressTarget := matcher.CombineTargetsIgnoringPrimaryKey(pod.Namespace, metav1.LabelSelector{MatchLabels: pod.Labels}, egressTargets)
		if combinedEgressTarget != nil {
			fmt.Printf("    combined egress:\n%s\n\n", strings.Join(matcher.ExplainTarget(combinedEgressTarget, egressValue), "\n"))
		} else {
			fmt.Println("    combined egress: none")
		}

		fmt.Printf("\n\n")
	}
}
