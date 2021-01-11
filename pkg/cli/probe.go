package cli

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/netpol/connectivity"
	"github.com/mattfenwick/cyclonus/pkg/netpol/matcher"
	"github.com/mattfenwick/cyclonus/pkg/netpol/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
)

type ProbeConnectivityArgs struct {
	PolicySource string
	Namespaces   []string
	PolicyPath   string
	ModelPath    string
}

type ProbeConnectivityConfig struct {
	Pods   *connectivity.PodModel
	Probes []*connectivity.ProtocolPort
}

func setupProbeConnectivityCommand() *cobra.Command {
	args := &ProbeConnectivityArgs{}

	command := &cobra.Command{
		Use:   "probe",
		Short: "probe connectivity",
		Long:  "probe connectivity against a cluster model; does not use a real cluster",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			runProbeConnectivityCommand(args)
		},
	}

	command.Flags().StringVar(&args.PolicySource, "policy-source", "kube", "source of network policies (kube, file, examples)")

	command.Flags().StringSliceVar(&args.Namespaces, "namespaces", []string{}, "only set if policy-source = kube; selects namespaces to read policies from; leaving empty will select all namespaces")

	command.Flags().StringVar(&args.PolicyPath, "policy-path", "", "only set if policy-source = file; path to network polic(ies)")

	command.Flags().StringVar(&args.ModelPath, "model-path", "", "path to json model file")
	utils.DoOrDie(command.MarkFlagRequired("model-path"))

	return command
}

func runProbeConnectivityCommand(args *ProbeConnectivityArgs) {
	// 1. source of policies
	kubePolicies, err := readPolicies(args.PolicySource, args.Namespaces, args.PolicyPath)
	utils.DoOrDie(err)

	// 2. consume policies
	explainedPolicies := matcher.BuildNetworkPolicies(kubePolicies)

	// 3. create config
	bs, err := ioutil.ReadFile(args.ModelPath)
	utils.DoOrDie(err)
	config := &ProbeConnectivityConfig{}
	err = json.Unmarshal(bs, &config)
	utils.DoOrDie(err)

	// 4. run probes
	for _, result := range connectivity.RunProbes(explainedPolicies, config.Probes, config.Pods) {
		log.Infof("probe on port %s, protocol %s", result.Port.Port.String(), result.Port.Protocol)

		// 5. print out a result matrix
		fmt.Println("Ingress:")
		result.Ingress.Table().Render()

		fmt.Println("Egress:")
		result.Egress.Table().Render()

		fmt.Println("Combined:")
		result.Combined.Table().Render()
	}
}
