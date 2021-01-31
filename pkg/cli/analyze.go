package cli

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/explainer"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/kube/netpol"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	networkingv1 "k8s.io/api/networking/v1"
)

type AnalyzePoliciesArgs struct {
	PolicySource string
	Namespaces   []string
	PolicyPath   string
	Format       string
	Context      string
}

func SetupAnalyzePoliciesCommand() *cobra.Command {
	args := &AnalyzePoliciesArgs{}

	command := &cobra.Command{
		Use:   "analyze",
		Short: "analyze network policies",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunAnalyzePoliciesCommand(args)
		},
	}

	command.Flags().StringVar(&args.PolicySource, "policy-source", "kube", "source of network policies (kube, file, examples)")

	command.Flags().StringSliceVar(&args.Namespaces, "namespaces", []string{}, "only set if policy-source = kube; selects namespaces to read policies from; leaving empty will select all namespaces")

	command.Flags().StringVar(&args.PolicyPath, "policy-path", "", "only set if policy-source = file; path to network polic(ies)")

	command.Flags().StringVar(&args.Format, "format", "table", "output format (options: json, table)")

	command.Flags().StringVar(&args.Context, "context", "", "only set if policy-source = kube; selects kube context to read policies from")

	return command
}

func RunAnalyzePoliciesCommand(args *AnalyzePoliciesArgs) {
	// 1. source of policies
	var kubePolicies []*networkingv1.NetworkPolicy
	var err error
	switch args.PolicySource {
	case "kube":
		kubeClient, err := kube.NewKubernetesForContext(args.Context)
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
	explainedPolicies := matcher.BuildNetworkPolicies(kubePolicies)
	switch args.Format {
	case "json":
		printJSON(explainedPolicies)
	case "table":
		explainer.TableExplainer(explainedPolicies).Render()
	default:
		fmt.Printf("%s\n\n", explainer.Explain(explainedPolicies))
	}
}
