package cli

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/synthetic"
	"github.com/mattfenwick/cyclonus/pkg/explainer"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/kube/netpol"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

type AnalyzeArgs struct {
	AllNamespaces      bool
	Namespaces         []string
	UseExamplePolicies bool
	PolicyPath         string
	Context            string

	// analyze
	Analyze bool

	// traffic
	TrafficPath string

	// targets
	TargetPodPath string

	// synthetic probe
	ProbePath string
}

func SetupAnalyzeCommand() *cobra.Command {
	args := &AnalyzeArgs{}

	command := &cobra.Command{
		Use:   "analyze",
		Short: "analyze network policies",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunAnalyzeCommand(args)
		},
	}

	command.Flags().BoolVar(&args.UseExamplePolicies, "use-example-policies", false, "if true, reads example policies")
	command.Flags().BoolVarP(&args.AllNamespaces, "all-namespaces", "A", true, "similar to kubectl's '--all-namespaces'/'-A' flag: if true, read policies from all-namespaces")
	command.Flags().StringSliceVarP(&args.Namespaces, "namespace", "n", []string{}, "similar to kubectl's '--namespace'/'-n' flag, except that multiple namespaces may be passed in; policies will be read from these namespaces")
	command.Flags().StringVar(&args.PolicyPath, "policy-path", "", "may be a file or a directory; if set, will attempt to read policies from the path")
	command.Flags().StringVar(&args.Context, "context", "", "only set if policy-source = kube; selects kube context to read policies from")

	command.Flags().BoolVar(&args.Analyze, "analyze", true, "if true, print analysis of network policies")
	command.Flags().StringVar(&args.TargetPodPath, "target-pod-path", "", "path to json target pod file -- json array of dicts; if empty, this step will be skipped")
	command.Flags().StringVar(&args.TrafficPath, "traffic-path", "", "path to json traffic file, containing of a list of traffic objects; if empty, this step will be skipped")
	command.Flags().StringVar(&args.ProbePath, "probe-path", "", "path to json model file for synthetic probe; if empty, this step will be skipped")

	return command
}

func RunAnalyzeCommand(args *AnalyzeArgs) {
	// 1. read policies from kube
	var kubePolicies []*networkingv1.NetworkPolicy
	namespaces := args.Namespaces
	if args.AllNamespaces {
		namespaces = []string{v1.NamespaceAll}
	}
	if len(namespaces) > 0 {
		kubeClient, err := kube.NewKubernetesForContext(args.Context)
		utils.DoOrDie(err)
		kubePolicies, err = readPoliciesFromKube(kubeClient, args.Namespaces)
	}
	// 2. read policies from file
	if args.PolicyPath != "" {
		policiesFromPath, err := readPoliciesFromPath(args.PolicyPath)
		utils.DoOrDie(err)
		kubePolicies = append(kubePolicies, policiesFromPath...)
	}
	// 3. read example policies
	if args.UseExamplePolicies {
		kubePolicies = append(kubePolicies, netpol.AllExamples...)
	}

	// 4. consume policies
	explainedPolicies := matcher.BuildNetworkPolicies(kubePolicies)

	if args.Analyze {
		AnalyzePolicies(explainedPolicies)
	}

	if args.TargetPodPath != "" {
		QueryTargets(explainedPolicies, args.TargetPodPath)
	}

	if args.TrafficPath != "" {
		QueryTraffic(explainedPolicies, args.TrafficPath)
	}

	if args.ProbePath != "" {
		ProbeSyntheticConnectivity(explainedPolicies, args.ProbePath)
	}
}

func AnalyzePolicies(explainedPolicies *matcher.Policy) {
	fmt.Printf("%s\n", explainer.TableExplainer(explainedPolicies))
}

type QueryTargetPod struct {
	Namespace string
	Labels    map[string]string
}

func QueryTargets(explainedPolicies *matcher.Policy, podPath string) {
	var pods []QueryTargetPod
	bs, err := ioutil.ReadFile(podPath)
	utils.DoOrDie(err)
	err = json.Unmarshal(bs, &pods)
	utils.DoOrDie(err)

	// TODO use tables for output
	for _, pod := range pods {
		fmt.Printf("pod %+v:\n\n", pod)

		// ingress
		fmt.Println("  ingress")
		ingressValue := true
		ingressTargets := explainedPolicies.TargetsApplyingToPod(ingressValue, pod.Namespace, pod.Labels)
		for _, t := range ingressTargets {
			fmt.Printf("    %s\n", strings.Join(explainer.ExplainTarget(t, ingressValue), "\n"))
		}
		// combine all the ingress targets for combined connectivity
		combinedIngressTarget := matcher.CombineTargetsIgnoringPrimaryKey(pod.Namespace, metav1.LabelSelector{MatchLabels: pod.Labels}, ingressTargets)
		if combinedIngressTarget != nil {
			fmt.Printf("    combined ingress:\n%s\n\n", strings.Join(explainer.ExplainTarget(combinedIngressTarget, ingressValue), "\n"))
		} else {
			fmt.Println("    combined ingress: none")
		}

		// egress
		fmt.Printf("\n  egress\n")
		egressValue := false
		egressTargets := explainedPolicies.TargetsApplyingToPod(egressValue, pod.Namespace, pod.Labels)
		for _, t := range egressTargets {
			fmt.Printf("    %s\n", strings.Join(explainer.ExplainTarget(t, egressValue), "\n"))
		}
		// combine all the egress targets for combined connectivity
		combinedEgressTarget := matcher.CombineTargetsIgnoringPrimaryKey(pod.Namespace, metav1.LabelSelector{MatchLabels: pod.Labels}, egressTargets)
		if combinedEgressTarget != nil {
			fmt.Printf("    combined egress:\n%s\n\n", strings.Join(explainer.ExplainTarget(combinedEgressTarget, egressValue), "\n"))
		} else {
			fmt.Println("    combined egress: none")
		}

		fmt.Printf("\n\n")
	}
}

func QueryTraffic(explainedPolicies *matcher.Policy, trafficPath string) {
	var allTraffics []*matcher.Traffic
	allTrafficBytes, err := ioutil.ReadFile(trafficPath)
	utils.DoOrDie(err)
	err = json.Unmarshal(allTrafficBytes, &allTraffics)
	utils.DoOrDie(err)
	for _, traffic := range allTraffics {
		trafficBytes, err := json.MarshalIndent(traffic, "", "  ")
		utils.DoOrDie(err)
		result := explainedPolicies.IsTrafficAllowed(traffic)
		fmt.Printf("Traffic:\n%s\n\nIs allowed: %t\n\n", string(trafficBytes), result.IsAllowed())
	}
}

type SyntheticProbeConnectivityConfig struct {
	Resources *synthetic.Resources
	Probes    []*struct {
		Protocol v1.Protocol
		Port     intstr.IntOrString
	}
}

func ProbeSyntheticConnectivity(explainedPolicies *matcher.Policy, modelPath string) {
	bs, err := ioutil.ReadFile(modelPath)
	utils.DoOrDie(errors.Wrapf(err, "unable to read file %s", modelPath))
	config := &SyntheticProbeConnectivityConfig{}
	err = json.Unmarshal(bs, &config)
	utils.DoOrDie(errors.Wrapf(err, "unable to unmarshal json"))

	// run probes
	for _, probe := range config.Probes {
		result := synthetic.RunSyntheticProbe(&synthetic.Request{
			Protocol:  probe.Protocol,
			Port:      probe.Port,
			Policies:  explainedPolicies,
			Resources: config.Resources,
		})

		logrus.Infof("probe on port %s, protocol %s", result.Request.Port.String(), result.Request.Protocol)

		// 5. print out a result matrix
		fmt.Println("Ingress:")
		fmt.Println(result.Ingress.Table())

		fmt.Println("Egress:")
		fmt.Println(result.Egress.Table())

		fmt.Println("Combined:")
		fmt.Println(result.Combined.Table())
	}
}
