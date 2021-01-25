package cli

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/kube/netpol"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
	networkingv1 "k8s.io/api/networking/v1"
)

type QueryTrafficArgs struct {
	PolicySource string
	Namespaces   []string
	TrafficPath  string
	PolicyPath   string
	Context      string
}

func SetupQueryTrafficCommand() *cobra.Command {
	args := &QueryTrafficArgs{}

	command := &cobra.Command{
		Use:   "traffic",
		Short: "query traffic allow/deny",
		Long:  "given policies and traffic as input, determine whether the traffic would be allowed",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunQueryTrafficCommand(args)
		},
	}

	command.Flags().StringVar(&args.PolicySource, "policy-source", "kube", "source of network policies (kube, file, examples)")

	command.Flags().StringSliceVar(&args.Namespaces, "namespaces", []string{}, "only set if policy-source = kube; selects namespaces to read policies from; leaving empty will select all namespaces")

	command.Flags().StringVar(&args.PolicyPath, "policy-path", "", "only set if policy-source = file; path to network polic(ies)")

	command.Flags().StringVar(&args.TrafficPath, "traffic-path", "", "path to traffic file, containing a list of traffic objects")
	utils.DoOrDie(command.MarkFlagRequired("traffic-path"))

	command.Flags().StringVar(&args.Context, "context", "", "only set if policy-source = kube; selects kube context to read policies from")

	return command
}

func RunQueryTrafficCommand(args *QueryTrafficArgs) {
	// 1. source of policies
	var kubePolicies []*networkingv1.NetworkPolicy
	var err error
	switch args.PolicySource {
	case "kube":
		var kubeClient *kube.Kubernetes
		if args.Context == "" {
			kubeClient, err = kube.NewKubernetesForDefaultContext()
		} else {
			kubeClient, err = kube.NewKubernetesForContext(args.Context)
		}
		utils.DoOrDie(err)
		kubePolicies, err = readPoliciesFromKube(kubeClient, args.Namespaces)
	case "file":
		kubePolicies, err = readPoliciesFromPath(args.PolicyPath)
	case "examples":
		kubePolicies = netpol.AllExamples
	default:
		panic(errors.Errorf("invalid policy source %s", args.PolicySource))
	}
	utils.DoOrDie(err)

	// 2. consume policies
	policies := matcher.BuildNetworkPolicies(kubePolicies)

	// 3. query
	var allTraffics []*matcher.Traffic
	allTrafficBytes, err := ioutil.ReadFile(args.TrafficPath)
	utils.DoOrDie(err)
	err = json.Unmarshal(allTrafficBytes, &allTraffics)
	utils.DoOrDie(err)
	for _, traffic := range allTraffics {
		trafficBytes, err := json.MarshalIndent(traffic, "", "  ")
		utils.DoOrDie(err)
		result := policies.IsTrafficAllowed(traffic)
		fmt.Printf("Traffic:\n%s\n\nIs allowed: %t\n\n", string(trafficBytes), result.IsAllowed())
	}
}
