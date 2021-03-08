package cli

import (
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	networkingv1 "k8s.io/api/networking/v1"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

func readPoliciesFromPath(policyPath string) ([]*networkingv1.NetworkPolicy, error) {
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
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "unable to read file %s", path)
		}

		// try parsing a list first
		var policies []*networkingv1.NetworkPolicy
		err = yaml.Unmarshal(bytes, &policies)
		if err == nil {
			log.Debugf("parsed %d policies from %s", len(policies), path)
			allPolicies = append(allPolicies, policies...)
			return nil
		}

		log.Debugf("failed to parse list from %s, falling back to parsing single policy", path)
		var policy *networkingv1.NetworkPolicy
		err = yaml.UnmarshalStrict(bytes, &policy)
		if err != nil {
			return errors.Wrapf(err, "unable to unmarshal single policy from yaml at %s", path)
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

func readPoliciesFromKube(kubeClient *kube.Kubernetes, namespaces []string) ([]*networkingv1.NetworkPolicy, error) {
	netpols, err := kube.GetNetworkPoliciesInNamespaces(kubeClient, namespaces)
	if err != nil {
		return nil, err
	}
	return refNetpolList(netpols), nil
}

func refNetpolList(refs []networkingv1.NetworkPolicy) []*networkingv1.NetworkPolicy {
	policies := make([]*networkingv1.NetworkPolicy, len(refs))
	for i := 0; i < len(refs); i++ {
		policies[i] = &refs[i]
	}
	return policies
}
