package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/kube/netpol/examples"
	"github.com/mattfenwick/cyclonus/pkg/netpol"
	"github.com/mattfenwick/cyclonus/pkg/netpol/explainer"
	"github.com/mattfenwick/cyclonus/pkg/netpol/matcher"
	"github.com/mattfenwick/cyclonus/pkg/netpol/utils"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"os"
)

func main() {
	explainerType := os.Args[1]
	ns := v1.NamespaceAll
	if len(os.Args) > 2 {
		ns = os.Args[2]
	}
	// sourceType := os.Args[3]

	// 1. source of policies
	var netpols *networkingv1.NetworkPolicyList
	if true {
		kubeClient, err := kube.NewKubernetes()
		utils.DoOrDie(err)
		netpols, err = kubeClient.ClientSet.NetworkingV1().NetworkPolicies(ns).List(context.TODO(), metav1.ListOptions{})
		utils.DoOrDie(err)
	} else {
		panic("TODO read some files?")
	}

	// 2. consume policies
	if explainerType == "explainer" {
		for _, policy := range netpols.Items {
			explanation := explainer.ExplainPolicy(&policy)
			printJSON(explanation)
		}
	} else {
		policies := make([]*networkingv1.NetworkPolicy, len(netpols.Items))
		for i := 0; i < len(netpols.Items); i++ {
			policies[i] = &netpols.Items[i]
		}
		explainedPolicies := matcher.BuildNetworkPolicies(policies)
		printJSON(explainedPolicies)
		fmt.Printf("%s\n\n", matcher.Explain(explainedPolicies))
	}

	if true {
		explainedPolicies := matcher.BuildNetworkPolicies(examples.AllExamples)
		printJSON(explainedPolicies)
		fmt.Printf("%s\n\n", matcher.Explain(explainedPolicies))
	}

	if false {
		mungeNetworkPolicies()
	}
}

// TODO connect
//func SetupNetpolCommand() *cobra.Command {
//	command := &cobra.Command{
//		Use:   "netpols",
//		Short: "netpol hacking",
//		Long:  "do stuff with network policies",
//		Args:  cobra.ExactArgs(0),
//		Run: func(cmd *cobra.Command, args []string) {
//			mungeNetworkPolicies()
//		},
//	}
//
//	return command
//}

func mungeNetworkPolicies() {
	k8s, err := kube.NewKubernetes()
	utils.DoOrDie(err)

	err = k8s.CleanNetworkPolicies("default")
	utils.DoOrDie(err)

	var allCreated []*networkingv1.NetworkPolicy
	for _, np := range examples.AllExamples {
		createdNp, err := k8s.CreateNetworkPolicy(np)
		allCreated = append(allCreated, createdNp)
		utils.DoOrDie(err)
		//explanation := netpol.ExplainPolicy(np)
		explanation := explainer.ExplainPolicy(createdNp)
		fmt.Printf("policy explanation for %s:\n%s\n\n", np.Name, explanation.PrettyPrint())

		matcherExplanation := matcher.Explain(matcher.BuildNetworkPolicy(createdNp))
		fmt.Printf("\nmatcher explanation: %s\n\n", matcherExplanation)

		reduced := netpol.Reduce(createdNp)
		fmt.Println(netpol.NodePrettyPrint(reduced))
		fmt.Println()

		fmt.Println("created netpol:")
		printJSON(createdNp)

		matcherPolicy := matcher.BuildNetworkPolicy(createdNp)
		matcherPolicyBytes, err := json.MarshalIndent(matcherPolicy, "", "  ")
		utils.DoOrDie(err)
		fmt.Printf("created matcher netpol:\n\n%s\n\n", matcherPolicyBytes)
		allowedResult := matcherPolicy.IsTrafficAllowed(&matcher.Traffic{
			Source: &matcher.TrafficPeer{
				Internal: &matcher.InternalPeer{
					NamespaceLabels: map[string]string{
						"app": "bookstore",
					},
					PodLabels: map[string]string{},
					Namespace: "not-default",
				},
				IP: "1.2.3.4",
			},
			Destination: &matcher.TrafficPeer{
				Internal: &matcher.InternalPeer{
					PodLabels: map[string]string{
						"app": "web",
					},
					NamespaceLabels: nil,
					Namespace:       "default",
				},
			},
			PortProtocol: &matcher.PortProtocol{
				Protocol: v1.ProtocolTCP,
				Port:     intstr.FromInt(9800),
			},
		})
		fmt.Printf("is allowed?  %t\n", allowedResult.IsAllowed())
		printJSON(allowedResult)
		fmt.Printf("\n\n")
	}

	netpols := matcher.BuildNetworkPolicies(allCreated)
	bytes, err := json.MarshalIndent(netpols, "", "  ")
	utils.DoOrDie(err)
	fmt.Printf("full network policies:\n\n%s\n\n", bytes)
	fmt.Printf("\nexplained:\n%s\n", matcher.Explain(netpols))

	netpolsExamples := matcher.BuildNetworkPolicy(examples.ExampleComplicatedNetworkPolicy())
	fmt.Printf("complicated example explained:\n%s\n", matcher.Explain(netpolsExamples))
}

func printJSON(obj interface{}) {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	utils.DoOrDie(err)
	fmt.Printf("%s\n", string(bytes))
}
