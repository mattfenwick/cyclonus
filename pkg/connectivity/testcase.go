package connectivity

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/explainer"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v12 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1"
	"sigs.k8s.io/yaml"
	"time"
)

type TestCase struct {
	KubePolicy                *v1.NetworkPolicy
	Noisy                     bool
	NetpolCreationWaitSeconds int
	Port                      int
	Protocol                  v12.Protocol
	PodModel                  *PodModel
	IgnoreLoopback            bool
	NamespacesToClean         []string
}

func TestNetworkPolicy(kubernetes *kube.Kubernetes, testCase *TestCase) error {
	utils.DoOrDie(kubernetes.DeleteAllNetworkPoliciesInNamespaces(testCase.NamespacesToClean))

	policy := matcher.BuildNetworkPolicy(testCase.KubePolicy)

	if testCase.Noisy {
		policyBytes, err := yaml.Marshal(testCase.KubePolicy)
		if err != nil {
			return errors.Wrapf(err, "unable to marshal network policy to yaml")
		}

		logrus.Infof("Creating network policy:\n%s\n\n", policyBytes)

		fmt.Printf("%s\n\n", explainer.Explain(policy))
		explainer.TableExplainer(policy).Render()
	}

	_, err := kubernetes.CreateNetworkPolicy(testCase.KubePolicy)
	if err != nil {
		return errors.Wrapf(err, "unable to marshal network policy to yaml")
	}

	logrus.Infof("waiting %d seconds for network policy to create and become active", testCase.NetpolCreationWaitSeconds)
	time.Sleep(time.Duration(testCase.NetpolCreationWaitSeconds) * time.Second)

	logrus.Infof("probe on port %d, protocol %s", testCase.Port, testCase.Protocol)
	synthetic := RunSyntheticProbe(policy, testCase.Protocol, testCase.Port, testCase.PodModel)

	kubeProbe := RunKubeProbe(kubernetes, testCase.PodModel, testCase.Port, testCase.Protocol, 5)

	fmt.Printf("\n\nKube results for %s/%s:\n", testCase.KubePolicy.Namespace, testCase.KubePolicy.Name)
	kubeProbe.Table().Render()

	comparison := synthetic.Combined.Compare(kubeProbe)
	t, f, nv, checked := comparison.ValueCounts(testCase.IgnoreLoopback)
	if f > 0 {
		fmt.Printf("Discrepancy found: %d wrong, %d no value, %d correct out of %d total\n", f, t, nv, checked)
	} else {
		fmt.Printf("found %d true, %d false, %d no value from %d total\n", t, f, nv, checked)
	}

	if f > 0 || testCase.Noisy {
		fmt.Println("Ingress:")
		synthetic.Ingress.Table().Render()

		fmt.Println("Egress:")
		synthetic.Egress.Table().Render()

		fmt.Println("Combined:")
		synthetic.Combined.Table().Render()

		fmt.Printf("\n\nSynthetic vs combined:\n")
		comparison.Table().Render()
	}

	return nil
}
