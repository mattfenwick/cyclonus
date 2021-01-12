package cli

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/netpol/connectivity"
	"github.com/mattfenwick/cyclonus/pkg/netpol/matcher"
	"github.com/mattfenwick/cyclonus/pkg/netpol/netpolgen"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

	panic(3)
	probes := []*connectivity.ProtocolPort{}
	podModel := connectivity.NewDefaultModel([]string{"x", "y", "z"}, []string{"a", "b", "c"})
	for _, kubeIngressPolicy := range kubePolicies {
		policy := matcher.BuildNetworkPolicy(kubeIngressPolicy)
		// 4. run probes
		for _, result := range connectivity.RunProbes(policy, probes, podModel) {
			log.Infof("probe on port %s, protocol %s", result.Port.Port.String(), result.Port.Protocol)

			// 5. print out a result matrix
			fmt.Println("Ingress:")
			result.Ingress.Table().Render()

			fmt.Println("Egress:")
			result.Egress.Table().Render()

			fmt.Println("Combined:")
			result.Combined.Table().Render()
		}

		// TODO run probes on cluster

		// TODO compare measure results to "expected" (calculated) results
	}
}
