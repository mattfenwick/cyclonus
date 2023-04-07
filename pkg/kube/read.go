package kube

import (
	"os"
	"path/filepath"

	"github.com/mattfenwick/collections/pkg/builtin"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	networkingv1 "k8s.io/api/networking/v1"
)

func ReadNetworkPoliciesFromPath(policyPath string) ([]*networkingv1.NetworkPolicy, error) {
	var allPolicies []*networkingv1.NetworkPolicy
	err := filepath.Walk(policyPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "unable to walk path %s", path)
		}
		if info.IsDir() {
			log.Tracef("not opening dir %s", path)
			return nil
		}
		log.Debugf("walking path %s", path)
		bytes, err := utils.ReadFileBytes(path)
		if err != nil {
			return err
		}

		// TODO try parsing plain yaml list (that is: not a NetworkPolicyList)
		// policies, err := utils.ParseYaml[[]*networkingv1.NetworkPolicy](bytes)

		// TODO try parsing multiple policies separated by '---' lines
		// policies, err := yaml.ParseMany[networkingv1.NetworkPolicy](bytes)
		// if err == nil {
		// 	log.Debugf("parsed %d policies from %s", len(policies), path)
		// 	allPolicies = append(allPolicies, refNetpolList(policies)...)
		// 	return nil
		// }
		// log.Errorf("unable to parse multiple policies separated by '---' lines: %+v", err)

		// try parsing a NetworkPolicyList
		policyList, err := utils.ParseYamlStrict[networkingv1.NetworkPolicyList](bytes)
		if err == nil {
			allPolicies = append(allPolicies, refNetpolList(policyList.Items)...)
			return nil
		}

		log.Debugf("unable to parse list of policies: %+v", err)

		policy, err := utils.ParseYamlStrict[networkingv1.NetworkPolicy](bytes)
		if err != nil {
			return errors.WithMessagef(err, "unable to parse single policy from yaml at %s", path)
		}

		log.Debugf("parsed single policy from %s: %+v", path, policy)
		allPolicies = append(allPolicies, policy)
		return nil
	})
	if err != nil {
		return nil, err
		//return nil, errors.Wrapf(err, "unable to walk filesystem from %s", policyPath)
	}
	for _, p := range allPolicies {
		if len(p.Spec.PolicyTypes) == 0 {
			return nil, errors.Errorf("missing spec.policyTypes from network policy %s/%s", p.Namespace, p.Name)
		}
	}
	return allPolicies, nil
}

func ReadNetworkPoliciesFromKube(kubeClient *Kubernetes, namespaces []string) ([]*networkingv1.NetworkPolicy, error) {
	netpols, err := GetNetworkPoliciesInNamespaces(kubeClient, namespaces)
	if err != nil {
		return nil, err
	}
	return refNetpolList(netpols), nil
}

func refNetpolList(refs []networkingv1.NetworkPolicy) []*networkingv1.NetworkPolicy {
	return slice.Map(builtin.Reference[networkingv1.NetworkPolicy], refs)
}
