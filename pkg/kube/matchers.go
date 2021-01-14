package kube

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"
)

// IsNameMatch follows the kube pattern of "empty string means matches All"
// It will return:
//   if matcher is empty: true
//   if objectName and matcher are the same: true
//   otherwise false
func IsNameMatch(objectName string, matcher string) bool {
	if matcher == "" {
		return true
	}
	return objectName == matcher
}

func IsMatchExpressionMatchForLabels(labels map[string]string, exp metav1.LabelSelectorRequirement) bool {
	switch exp.Operator {
	case metav1.LabelSelectorOpIn:
		val, ok := labels[exp.Key]
		if !ok {
			return false
		}
		for _, v := range exp.Values {
			if v == val {
				return true
			}
		}
		return false
	case metav1.LabelSelectorOpNotIn:
		val, ok := labels[exp.Key]
		if !ok {
			// see https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#resources-that-support-set-based-requirements
			//   even for NotIn -- if the key isn't there, it's not a match
			return false
		}
		for _, v := range exp.Values {
			if v == val {
				return false
			}
		}
		return true
	case metav1.LabelSelectorOpExists:
		_, ok := labels[exp.Key]
		return ok
	case metav1.LabelSelectorOpDoesNotExist:
		_, ok := labels[exp.Key]
		return !ok
	default:
		panic("invalid operator")
	}
}

// IsLabelsMatchLabelSelector matches labels to a kube LabelSelector.
// From the docs:
// > A label selector is a label query over a set of resources. The result of matchLabels and
// > matchExpressions are ANDed. An empty label selector matches all objects. A null
// > label selector matches no objects.
func IsLabelsMatchLabelSelector(labels map[string]string, labelSelector metav1.LabelSelector) bool {
	// From the docs: "The requirements are ANDed."
	//   Therefore, all MatchLabels must be matched.
	for key, val := range labelSelector.MatchLabels {
		if labels[key] != val {
			return false
		}
	}

	// From the docs: "The requirements are ANDed."
	//   Therefore, all MatchExpressions must be matched.
	for _, exp := range labelSelector.MatchExpressions {
		isMatch := IsMatchExpressionMatchForLabels(labels, exp)
		if !isMatch {
			return false
		}
	}

	// From the docs: "An empty label selector matches all objects."
	return true
}

func IsIPInCIDR(ip string, cidr string) (bool, error) {
	_, cidrNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, errors.Wrapf(err, "unable to parse CIDR '%s'", cidr)
	}
	trafficIP := net.ParseIP(ip)
	if trafficIP == nil {
		return false, errors.Errorf("unable to parse IP '%s'", ip)
	}
	return cidrNet.Contains(trafficIP), nil
}

func IsIPAddressMatchForIPBlock(ip string, ipBlock *v1.IPBlock) (bool, error) {
	isInCidr, err := IsIPInCIDR(ip, ipBlock.CIDR)
	if err != nil {
		return false, err
	}
	if !isInCidr {
		return false, nil
	}
	for _, except := range ipBlock.Except {
		isInExcept, err := IsIPInCIDR(ip, except)
		if err != nil {
			return false, err
		}
		if isInExcept {
			return false, nil
		}
	}
	return true, nil
}
