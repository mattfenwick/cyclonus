package cli

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/netpol/connectivity"
	"github.com/mattfenwick/cyclonus/pkg/netpol/matcher"
	"github.com/mattfenwick/cyclonus/pkg/netpol/netpolgen"
	"github.com/mattfenwick/cyclonus/pkg/netpol/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"time"
)

type GeneratorArgs struct {
}

func setupGeneratorCommand() *cobra.Command {
	args := &GeneratorArgs{}

	command := &cobra.Command{
		Use:   "generator",
		Short: "generate network policies",
		Long:  "generate network policies including corner cases by combinations of fragments",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			runGeneratorCommand(args)
		},
	}

	return command
}

func runGeneratorCommand(args *GeneratorArgs) {
	kubePolicies := netpolgen.NewDefaultGenerator().IngressPolicies()
	fmt.Printf("%d policies\n\n", len(kubePolicies))

	port := &connectivity.ProtocolPort{
		Protocol: v1.ProtocolTCP,
		Port:     80,
	}
	// TODO don't use default model?
	namespaces := []string{"x", "y", "z"}
	podModel := connectivity.NewDefaultModel(namespaces, []string{"a", "b", "c"}, port.Port, port.Protocol)
	kubernetes, err := kube.NewKubernetes()
	utils.DoOrDie(err)

	utils.DoOrDie(connectivity.CreateResources(kubernetes, podModel))
	// TODO wait for pods to come up
	time.Sleep(10 * time.Second)

	namespacesToCleanSet := map[string]bool{}
	namespacesToClean := []string{}
	for _, kp := range kubePolicies {
		if !namespacesToCleanSet[kp.Namespace] {
			namespacesToCleanSet[kp.Namespace] = true
			namespacesToClean = append(namespacesToClean, kp.Namespace)
		}
	}
	for _, ns := range namespaces {
		if !namespacesToCleanSet[ns] {
			namespacesToCleanSet[ns] = true
			namespacesToClean = append(namespacesToClean, ns)
		}
	}

	for i, kubeIngressPolicy := range kubePolicies {
		utils.DoOrDie(kubernetes.DeleteAllNetworkPoliciesInNamespaces(namespacesToClean))

		_, err = kubernetes.CreateNetworkPolicy(kubeIngressPolicy)
		utils.DoOrDie(err)

		// TODO wait for netpol to become 'active'
		time.Sleep(1 * time.Second)

		policy := matcher.BuildNetworkPolicy(kubeIngressPolicy)

		log.Infof("probe on port %d, protocol %s", port.Port, port.Protocol)
		synthetic := connectivity.RunSyntheticProbe(policy, port, podModel)
		fmt.Println("Ingress:")
		synthetic.Ingress.Table().Render()
		fmt.Println("Egress:")
		synthetic.Egress.Table().Render()
		fmt.Println("Combined:")
		synthetic.Combined.Table().Render()

		kubeProbe := connectivity.RunKubeProbe(kubernetes, podModel, port.Port, port.Protocol, 5)
		fmt.Printf("\n\nKube results:\n")
		kubeProbe.Table().Render()

		fmt.Printf("\n\nSynthetic vs combined:\n")
		synthetic.Combined.Compare(kubeProbe).Table().Render()

		fmt.Println()

		if i > 2 {
			panic(i)
		}
	}
}
