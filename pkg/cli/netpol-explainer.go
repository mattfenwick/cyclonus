package cli

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/netpol/matcher"
	"github.com/mattfenwick/cyclonus/pkg/netpol/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func RunRootCommand() {
	command := setupRootCommand()
	if err := errors.Wrapf(command.Execute(), "run root command"); err != nil {
		log.Fatalf("unable to run root command: %+v", err)
		os.Exit(1)
	}
}

type Flags struct {
	Verbosity string
}

func setupRootCommand() *cobra.Command {
	flags := &Flags{}
	command := &cobra.Command{
		Use:   "netpol-explainer",
		Short: "explain, analyze, and query network policies",
		Long:  "explain, analyze, and query network policies",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return SetUpLogger(flags.Verbosity)
		},
	}

	command.PersistentFlags().StringVarP(&flags.Verbosity, "verbosity", "v", "info", "log level; one of [info, debug, trace, warn, error, fatal, panic]")

	command.AddCommand(setupAnalyzePoliciesCommand())
	command.AddCommand(setupQueryTrafficCommand())
	command.AddCommand(setupProbeConnectivityCommand())

	// TODO
	//command.AddCommand(setupQueryTargetsCommand())
	//command.AddCommand(setupQueryPeersCommand())

	return command
}

type AnalyzePoliciesArgs struct {
	PolicySource string
	Namespaces   []string
	PolicyPath   string
	Format       string
}

func setupAnalyzePoliciesCommand() *cobra.Command {
	args := &AnalyzePoliciesArgs{}

	command := &cobra.Command{
		Use:   "analyze",
		Short: "analyze network policies",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			runAnalyzePoliciesCommand(args)
		},
	}

	command.Flags().StringVar(&args.PolicySource, "policy-source", "kube", "source of network policies (kube, file, examples)")

	command.Flags().StringSliceVar(&args.Namespaces, "namespaces", []string{}, "only set if policy-source = kube; selects namespaces to read policies from; leaving empty will select all namespaces")

	command.Flags().StringVar(&args.PolicyPath, "policy-path", "", "only set if policy-source = file; path to network polic(ies)")

	command.Flags().StringVar(&args.Format, "format", "", "output format; human-readable if empty (options: json)")

	return command
}

func runAnalyzePoliciesCommand(args *AnalyzePoliciesArgs) {
	// 1. source of policies
	kubePolicies, err := readPolicies(args.PolicySource, args.Namespaces, args.PolicyPath)
	utils.DoOrDie(err)

	// 2. consume policies
	explainedPolicies := matcher.BuildNetworkPolicies(kubePolicies)
	switch args.Format {
	case "json":
		printJSON(explainedPolicies)
	default:
		fmt.Printf("%s\n\n", matcher.Explain(explainedPolicies))
	}
}
