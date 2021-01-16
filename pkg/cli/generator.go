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
	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/yaml"
	"time"
)

type GeneratorArgs struct {
	Ingress        bool
	Egress         bool
	AllowDNS       bool
	PrintSynthetic bool
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

	command.Flags().BoolVar(&args.Ingress, "ingress", true, "use ingress in policies")
	command.Flags().BoolVar(&args.Egress, "egress", false, "use egress in policies")
	command.Flags().BoolVar(&args.AllowDNS, "allow-dns", true, "if using egress, allow udp over port 53 for DNS resolution")
	command.Flags().BoolVar(&args.PrintSynthetic, "print-synthetic", false, "if true, print synthetic results")

	return command
}

func runGeneratorCommand(args *GeneratorArgs) {
	namespaces := []string{"x", "y", "z"}
	generator := &netpolgen.Generator{
		Ports:            netpolgen.DefaultPorts(),
		Peers:            netpolgen.DefaultPeers(),
		Targets:          netpolgen.DefaultTargets(),
		Namespaces:       namespaces,
		TypicalPorts:     netpolgen.TypicalPorts,
		TypicalPeers:     netpolgen.TypicalPeers,
		TypicalTarget:    netpolgen.TypicalTarget,
		TypicalNamespace: netpolgen.TypicalNamespace,
	}

	var kubePolicies []*networkingv1.NetworkPolicy
	if args.Ingress && args.Egress {
		kubePolicies = generator.IngressEgressPolicies(args.AllowDNS)
	} else if args.Ingress {
		kubePolicies = generator.IngressPolicies()
	} else if args.Egress {
		kubePolicies = generator.EgressPolicies(args.AllowDNS)
	} else {
		// TODO clean this up
		//panic(errors.Errorf("no policies to test: neither ingress nor egress allowed: %+v", args))

		// ingress
		//kubePolicies = generator.VaryIngressPolicies()

		// egress
		kubePolicies = generator.VaryEgressPolicies(args.AllowDNS)
	}
	fmt.Printf("testing %d policies\n\n", len(kubePolicies))

	port := &connectivity.ProtocolPort{
		Protocol: v1.ProtocolTCP,
		Port:     80,
	}
	// TODO don't use default model?
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

	for i, kubePolicy := range kubePolicies {
		utils.DoOrDie(kubernetes.DeleteAllNetworkPoliciesInNamespaces(namespacesToClean))

		_, err = kubernetes.CreateNetworkPolicy(kubePolicy)
		utils.DoOrDie(err)

		// TODO wait for netpol to become 'active'
		time.Sleep(1 * time.Second)

		policy := matcher.BuildNetworkPolicy(kubePolicy)

		log.Infof("probe on port %d, protocol %s", port.Port, port.Protocol)
		synthetic := connectivity.RunSyntheticProbe(policy, port, podModel)

		kubeProbe := connectivity.RunKubeProbe(kubernetes, podModel, port.Port, port.Protocol, 5)

		fmt.Printf("\n\nKube results for %s/%s:\n", kubePolicy.Namespace, kubePolicy.Name)
		kubeProbe.Table().Render()

		comparison := synthetic.Combined.Compare(kubeProbe)
		t, f, nv := comparison.ValueCounts()
		log.Infof("found %d true, %d false, %d no value", t, f, nv)
		if f > 0 {
			policyBytes, err := yaml.Marshal(kubePolicy)
			utils.DoOrDie(err)
			fmt.Printf("Discrepancy found for network policy:\n%s\n\n", policyBytes)
		}

		if f > 0 || args.PrintSynthetic {
			fmt.Println("Ingress:")
			synthetic.Ingress.Table().Render()

			fmt.Println("Egress:")
			synthetic.Egress.Table().Render()

			fmt.Println("Combined:")
			synthetic.Combined.Table().Render()

			fmt.Printf("\n\nSynthetic vs combined:\n")
			comparison.Table().Render()
		}

		fmt.Printf("\nfinished policy #%d\n\n", i)
	}
}
