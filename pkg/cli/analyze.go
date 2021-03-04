package cli

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/probe"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/linter"
	"io/ioutil"
	"strings"

	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/kube/netpol"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ExplainMode      = "explain"
	LintMode         = "lint"
	QueryTrafficMode = "query-traffic"
	QueryTargetMode  = "query-target"
	ProbeMode        = "probe"
)

var AllModes = []string{
	ExplainMode,
	LintMode,
	QueryTrafficMode,
	QueryTargetMode,
	ProbeMode,
}

type AnalyzeArgs struct {
	AllNamespaces      bool
	Namespaces         []string
	UseExamplePolicies bool
	PolicyPath         string
	Context            string

	Modes []string

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
	command.Flags().BoolVarP(&args.AllNamespaces, "all-namespaces", "A", false, "reads kube resources from all namespaces; same as kubectl's '--all-namespaces'/'-A' flag")
	command.Flags().StringSliceVarP(&args.Namespaces, "namespace", "n", []string{"default"}, "namespaces to read kube resources from ; same as kubectl's '--namespace'/'-n' flag, except that multiple namespaces may be passed in")
	command.Flags().StringVar(&args.PolicyPath, "policy-path", "", "may be a file or a directory; if set, will attempt to read policies from the path")
	command.Flags().StringVar(&args.Context, "context", "", "selects kube context to read policies from; only reads from kube if one or more namespaces or all namespaces are specified")

	command.Flags().StringSliceVar(&args.Modes, "mode", []string{ExplainMode}, "analysis modes to run; allowed values are "+strings.Join(AllModes, ","))

	command.Flags().StringVar(&args.TargetPodPath, "target-pod-path", "", "path to json target pod file -- json array of dicts")
	command.Flags().StringVar(&args.TrafficPath, "traffic-path", "", "path to json traffic file, containing of a list of traffic objects")
	command.Flags().StringVar(&args.ProbePath, "probe-path", "", "path to json model file for synthetic probe")

	return command
}

func RunAnalyzeCommand(args *AnalyzeArgs) {
	// 1. read policies from kube
	var kubePolicies []*networkingv1.NetworkPolicy
	var kubePods []v1.Pod
	var kubeNamespaces []v1.Namespace
	if args.AllNamespaces || len(args.Namespaces) > 0 {
		kubeClient, err := kube.NewKubernetesForContext(args.Context)
		utils.DoOrDie(err)

		namespaces := args.Namespaces
		if args.AllNamespaces {
			nsList, err := kubeClient.GetAllNamespaces()
			utils.DoOrDie(err)
			kubeNamespaces = nsList.Items
			namespaces = []string{v1.NamespaceAll}
		}
		kubePolicies, err = readPoliciesFromKube(kubeClient, namespaces)
		kubePods, err = kubeClient.GetPodsInNamespaces(namespaces)
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

	logrus.Debugf("parsed policies:\n%s", utils.JsonString(kubePolicies))

	// 4. consume policies
	explainedPolicies := matcher.BuildNetworkPolicies(kubePolicies)

	for _, mode := range args.Modes {
		switch mode {
		case ExplainMode:
			ExplainPolicies(explainedPolicies)
		case LintMode:
			Lint(kubePolicies)
		case QueryTargetMode:
			pods := make([]*QueryTargetPod, len(kubePods))
			for i, p := range kubePods {
				pods[i] = &QueryTargetPod{
					Namespace: p.Namespace,
					Labels:    p.Labels,
				}
			}
			QueryTargets(explainedPolicies, args.TargetPodPath, pods)
		case QueryTrafficMode:
			QueryTraffic(explainedPolicies, args.TrafficPath)
		case ProbeMode:
			ProbeSyntheticConnectivity(explainedPolicies, args.ProbePath, kubePods, kubeNamespaces)
		default:
			panic(errors.Errorf("unrecognized mode %s", mode))
		}
	}
}

func ExplainPolicies(explainedPolicies *matcher.Policy) {
	fmt.Printf("%s\n", explainedPolicies.ExplainTable())
}

func Lint(kubePolicies []*networkingv1.NetworkPolicy) {
	warnings := linter.Lint(kubePolicies, map[linter.Check]bool{})
	fmt.Println(linter.WarningsTable(warnings))
}

// QueryTargetPod matches targets; targets exist in only a single namespace and can't be matched by namespace
//   label, therefore we match by exact namespace and by pod labels.
type QueryTargetPod struct {
	Namespace string
	Labels    map[string]string
}

func QueryTargets(explainedPolicies *matcher.Policy, podPath string, pods []*QueryTargetPod) {
	if podPath != "" {
		var podsFromFile []*QueryTargetPod
		bs, err := ioutil.ReadFile(podPath)
		utils.DoOrDie(err)
		err = json.Unmarshal(bs, &podsFromFile)
		utils.DoOrDie(err)
		pods = append(pods, podsFromFile...)
	}

	for _, pod := range pods {
		fmt.Printf("pod in ns %s with labels %+v:\n\n", pod.Namespace, pod.Labels)

		targets, combinedRules := QueryTargetHelper(explainedPolicies, pod)

		fmt.Printf("Matching targets:\n%s\n", targets.ExplainTable())
		fmt.Printf("Combined rules:\n%s\n\n\n", combinedRules.ExplainTable())
	}
}

func QueryTargetHelper(policies *matcher.Policy, pod *QueryTargetPod) (*matcher.Policy, *matcher.Policy) {
	ingressTargets := policies.TargetsApplyingToPod(true, pod.Namespace, pod.Labels)
	combinedIngressTarget := matcher.CombineTargetsIgnoringPrimaryKey(pod.Namespace, metav1.LabelSelector{MatchLabels: pod.Labels}, ingressTargets)

	egressTargets := policies.TargetsApplyingToPod(false, pod.Namespace, pod.Labels)
	combinedEgressTarget := matcher.CombineTargetsIgnoringPrimaryKey(pod.Namespace, metav1.LabelSelector{MatchLabels: pod.Labels}, egressTargets)

	var combinedIngresses []*matcher.Target
	if combinedIngressTarget != nil {
		combinedIngresses = []*matcher.Target{combinedIngressTarget}
	}
	var combinedEgresses []*matcher.Target
	if combinedEgressTarget != nil {
		combinedEgresses = []*matcher.Target{combinedEgressTarget}
	}

	return matcher.NewPolicyWithTargets(ingressTargets, egressTargets), matcher.NewPolicyWithTargets(combinedIngresses, combinedEgresses)

}

func QueryTraffic(explainedPolicies *matcher.Policy, trafficPath string) {
	var allTraffics []*matcher.Traffic
	if trafficPath == "" {
		logrus.Fatalf("%+v", errors.Errorf("path to traffic file required for QueryTraffic command"))
	}
	allTrafficBytes, err := ioutil.ReadFile(trafficPath)
	utils.DoOrDie(err)
	err = json.Unmarshal(allTrafficBytes, &allTraffics)
	utils.DoOrDie(err)

	for _, traffic := range allTraffics {
		fmt.Printf("Traffic:\n%s\n", traffic.Table())

		result := explainedPolicies.IsTrafficAllowed(traffic)
		fmt.Printf("Is traffic allowed?\n%s\n\n\n", result.Table())
	}
}

type SyntheticProbeConnectivityConfig struct {
	Resources *probe.Resources
	Probes    []*generator.PortProtocol
}

func ProbeSyntheticConnectivity(explainedPolicies *matcher.Policy, modelPath string, kubePods []v1.Pod, kubeNamespaces []v1.Namespace) {
	if modelPath != "" {
		bs, err := ioutil.ReadFile(modelPath)
		utils.DoOrDie(errors.Wrapf(err, "unable to read file %s", modelPath))
		config := &SyntheticProbeConnectivityConfig{}
		err = json.Unmarshal(bs, &config)
		utils.DoOrDie(errors.Wrapf(err, "unable to unmarshal json"))

		// run probes
		for _, probeConfig := range config.Probes {
			probeResult := probe.NewSimulatedRunner(explainedPolicies).
				RunProbeFixedPortProtocol(config.Resources, probeConfig.Port, probeConfig.Protocol)

			logrus.Infof("probe on port %s, protocol %s", probeConfig.Port.String(), probeConfig.Protocol)

			fmt.Printf("Ingress:\n%s\n", probeResult.RenderIngress())

			fmt.Printf("Egress:\n%s\n", probeResult.RenderEgress())

			fmt.Printf("Combined:\n%s\n\n\n", probeResult.RenderTable())
		}
	}

	resources := &probe.Resources{
		Namespaces: map[string]map[string]string{},
		Pods:       []*probe.Pod{},
	}

	nsMap := map[string]v1.Namespace{}
	for _, ns := range kubeNamespaces {
		nsMap[ns.Name] = ns
		resources.Namespaces[ns.Name] = ns.Labels
	}

	for _, pod := range kubePods {
		var containers []*probe.Container
		for _, cont := range pod.Spec.Containers {
			if len(cont.Ports) == 0 {
				logrus.Warnf("skipping container %s/%s/%s, no ports available", pod.Namespace, pod.Name, cont.Name)
				continue
			}
			port := cont.Ports[0]
			containers = append(containers, &probe.Container{
				Name:     cont.Name,
				Port:     int(port.ContainerPort),
				Protocol: port.Protocol,
				PortName: port.Name,
			})
		}
		if len(containers) == 0 {
			logrus.Warnf("skipping pod %s/%s, no containers available", pod.Namespace, pod.Name)
			continue
		}
		resources.Pods = append(resources.Pods, &probe.Pod{
			Namespace:  pod.Namespace,
			Name:       pod.Name,
			Labels:     pod.Labels,
			IP:         pod.Status.PodIP,
			Containers: containers,
		})
	}

	simRunner := probe.NewSimulatedRunner(explainedPolicies)
	simulatedProbe := simRunner.RunProbeForConfig(generator.ProbeAllAvailable, resources)
	fmt.Printf("Ingress:\n%s\n", simulatedProbe.RenderIngress())
	fmt.Printf("Egress:\n%s\n", simulatedProbe.RenderEgress())
	fmt.Printf("Combined:\n%s\n\n\n", simulatedProbe.RenderTable())
}
