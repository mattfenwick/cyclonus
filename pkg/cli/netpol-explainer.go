package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/kube/netpol/examples"
	"github.com/mattfenwick/cyclonus/pkg/netpol/matcher"
	"github.com/mattfenwick/cyclonus/pkg/netpol/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// TODO
	//command.AddCommand(setupQueryTargetsCommand())
	//command.AddCommand(setupQueryPeersCommand())
	//command.AddCommand(setupQueryTrafficCommand())

	return command
}

type AnalyzePoliciesArgs struct {
	PolicySource string
	Namespaces   []string
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

	return command
}

func runAnalyzePoliciesCommand(args *AnalyzePoliciesArgs) {
	// 1. source of policies
	kubePolicies, err := readPolicies(args.PolicySource, args.Namespaces)
	utils.DoOrDie(err)

	// 2. consume policies
	explainedPolicies := matcher.BuildNetworkPolicies(kubePolicies)
	printJSON(explainedPolicies)
	fmt.Printf("%s\n\n", matcher.Explain(explainedPolicies))
}

func readPolicies(source string, namespaces []string) ([]*networkingv1.NetworkPolicy, error) {
	switch source {
	case "kube":
		return readPoliciesFromKube(namespaces)
	case "file":
		return nil, errors.Errorf("TODO -- unimplemented")
	case "examples":
		return examples.AllExamples, nil
	default:
		return nil, errors.Errorf("invalid policy source %s", source)
	}
}

func readPoliciesFromKube(namespaces []string) ([]*networkingv1.NetworkPolicy, error) {
	kubeClient, err := kube.NewKubernetes()
	if err != nil {
		return nil, err
	}
	if len(namespaces) == 0 {
		list, err := kubeClient.ClientSet.NetworkingV1().NetworkPolicies(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to list netpols in all namespaces")
		}
		return refNetpolList(list.Items), nil
	} else {
		var list []*networkingv1.NetworkPolicy
		for _, ns := range namespaces {
			nsList, err := kubeClient.ClientSet.NetworkingV1().NetworkPolicies(ns).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return nil, errors.Wrapf(err, "unable to list netpols in namespace %s", ns)
			}
			list = append(list, refNetpolList(nsList.Items)...)
		}
		return list, nil
	}
}

func refNetpolList(refs []networkingv1.NetworkPolicy) []*networkingv1.NetworkPolicy {
	policies := make([]*networkingv1.NetworkPolicy, len(refs))
	for i := 0; i < len(refs); i++ {
		policies[i] = &refs[i]
	}
	return policies
}

func printJSON(obj interface{}) {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	utils.DoOrDie(err)
	fmt.Printf("%s\n", string(bytes))
}

func SetUpLogger(logLevelStr string) error {
	logLevel, err := log.ParseLevel(logLevelStr)
	if err != nil {
		return errors.Wrapf(err, "unable to parse the specified log level: '%s'", logLevel)
	}
	log.SetLevel(logLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.Infof("log level set to '%s'", log.GetLevel())
	return nil
}
