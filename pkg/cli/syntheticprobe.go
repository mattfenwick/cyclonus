package cli

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/synthetic"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/kube/netpol"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

type SyntheticProbeConnectivityArgs struct {
	PolicySource string
	Namespaces   []string
	PolicyPath   string
	ModelPath    string
	Context      string
}

type SyntheticProbeConnectivityConfig struct {
	Pods   *synthetic.Resources
	Probes []*struct {
		Protocol v1.Protocol
		Port     int
	}
}

func SetupSyntheticProbeConnectivityCommand() *cobra.Command {
	args := &SyntheticProbeConnectivityArgs{}

	command := &cobra.Command{
		Use:   "synthetic-probe",
		Short: "probe synthetic connectivity",
		Long:  "probe connectivity against a cluster model; does not use a real cluster",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunProbeSyntheticConnectivityCommand(args)
		},
	}

	command.Flags().StringVar(&args.PolicySource, "policy-source", "kube", "source of network policies (kube, file, examples)")

	command.Flags().StringSliceVar(&args.Namespaces, "namespaces", []string{}, "only set if policy-source = kube; selects namespaces to read policies from; leaving empty will select all namespaces")

	command.Flags().StringVar(&args.PolicyPath, "policy-path", "", "only set if policy-source = file; path to network polic(ies)")

	command.Flags().StringVar(&args.ModelPath, "model-path", "", "path to json model file")
	utils.DoOrDie(command.MarkFlagRequired("model-path"))

	command.Flags().StringVar(&args.Context, "context", "", "only set if policy-source = kube; selects kube context to read policies from")

	return command
}

func RunProbeSyntheticConnectivityCommand(args *SyntheticProbeConnectivityArgs) {
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
	explainedPolicies := matcher.BuildNetworkPolicies(kubePolicies)

	// 3. create config
	bs, err := ioutil.ReadFile(args.ModelPath)
	utils.DoOrDie(errors.Wrapf(err, "unable to read file %s", args.ModelPath))
	config := &SyntheticProbeConnectivityConfig{}
	err = json.Unmarshal(bs, &config)
	utils.DoOrDie(errors.Wrapf(err, "unable to unmarshal json"))

	// 4. run probes
	for _, probe := range config.Probes {
		result := synthetic.RunSyntheticProbe(&synthetic.Request{
			Protocol:  probe.Protocol,
			Port:      probe.Port,
			Policies:  explainedPolicies,
			Resources: config.Pods,
		})

		log.Infof("probe on port %d, protocol %s", result.Request.Port, result.Request.Protocol)

		// 5. print out a result matrix
		fmt.Println("Ingress:")
		result.Ingress.Table().Render()

		fmt.Println("Egress:")
		result.Egress.Table().Render()

		fmt.Println("Combined:")
		result.Combined.Table().Render()
	}
}
