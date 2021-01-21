package cli

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/explainer"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/spf13/cobra"
)

type AnalyzePoliciesArgs struct {
	PolicySource string
	Namespaces   []string
	PolicyPath   string
	Format       string
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

	return command
}

func RunAnalyzePoliciesCommand(args *AnalyzePoliciesArgs) {
	// 1. source of policies
	kubePolicies, err := readPolicies(args.PolicySource, args.Namespaces, args.PolicyPath)
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
