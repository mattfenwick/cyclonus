package cli

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/netpol/matcher"
	"github.com/mattfenwick/cyclonus/pkg/netpol/utils"
	"github.com/spf13/cobra"
	"io/ioutil"
)

type QueryTrafficArgs struct {
	PolicySource string
	Namespaces   []string
	TrafficFile  string
	PolicyPath   string
}

func setupQueryTrafficCommand() *cobra.Command {
	args := &QueryTrafficArgs{}

	command := &cobra.Command{
		Use:   "traffic",
		Short: "query traffic allow/deny",
		Long:  "given policies and traffic as input, determine whether the traffic would be allowed",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			runQueryTrafficCommand(args)
		},
	}

	command.Flags().StringVar(&args.PolicySource, "policy-source", "kube", "source of network policies (kube, file, examples)")

	command.Flags().StringSliceVar(&args.Namespaces, "namespaces", []string{}, "only set if policy-source = kube; selects namespaces to read policies from; leaving empty will select all namespaces")

	command.Flags().StringVar(&args.PolicyPath, "policy-path", "", "only set if policy-source = file; path to network polic(ies)")

	command.Flags().StringVar(&args.TrafficFile, "traffic-file", "", "path to traffic file, containing a list of traffic objects")
	command.MarkFlagRequired("traffic-file")

	return command
}

func runQueryTrafficCommand(args *QueryTrafficArgs) {
	// 1. source of policies
	kubePolicies, err := readPolicies(args.PolicySource, args.Namespaces, args.PolicyPath)
	utils.DoOrDie(err)

	// 2. consume policies
	explainedPolicies := matcher.BuildNetworkPolicies(kubePolicies)

	// 3. query
	var allTraffics []*matcher.Traffic
	allTrafficBytes, err := ioutil.ReadFile(args.TrafficFile)
	utils.DoOrDie(err)
	err = json.Unmarshal(allTrafficBytes, &allTraffics)
	utils.DoOrDie(err)
	for _, traffic := range allTraffics {
		trafficBytes, err := json.MarshalIndent(traffic, "", "  ")
		utils.DoOrDie(err)
		result := explainedPolicies.IsTrafficAllowed(traffic)
		resultBytes, err := json.MarshalIndent(result, "", "  ")
		utils.DoOrDie(err)
		fmt.Printf("Traffic:\n%s\n\nIs allowed: %t\n\nExplanation:\n\n%s\n\n", string(trafficBytes), result.IsAllowed(), string(resultBytes))
	}
}
