package kube

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/slices"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"golang.org/x/exp/maps"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
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

func IsLabelSelectorEmpty(l metav1.LabelSelector) bool {
	return len(l.MatchLabels) == 0 && len(l.MatchExpressions) == 0
}

// SerializeLabelSelector deterministically converts a metav1.LabelSelector
// into a string
func SerializeLabelSelector(ls metav1.LabelSelector) string {
	keyVals := slices.Map(func(key string) string {
		return fmt.Sprintf("%s: %s", key, ls.MatchLabels[key])
	}, slices.Sort(maps.Keys(ls.MatchLabels)))
	// this looks weird -- we're using an array to make the order deterministic
	return utils.JsonStringNoIndent([]interface{}{"MatchLabels", keyVals, "MatchExpression", ls.MatchExpressions})
}

func LabelSelectorTableLines(selector metav1.LabelSelector) string {
	if IsLabelSelectorEmpty(selector) {
		return "all pods"
	}
	var lines []string
	if len(selector.MatchLabels) > 0 {
		lines = append(lines, "Match labels:")
		for _, key := range slices.Sort(maps.Keys(selector.MatchLabels)) {
			val := selector.MatchLabels[key]
			lines = append(lines, fmt.Sprintf("  %s: %s", key, val))
		}
	}
	if len(selector.MatchExpressions) > 0 {
		lines = append(lines, "Match expressions:")
		sortedMatchExpressions := slices.SortOn(
			func(l metav1.LabelSelectorRequirement) string { return l.Key },
			selector.MatchExpressions)
		for _, exp := range sortedMatchExpressions {
			lines = append(lines, fmt.Sprintf("  %s %s %+v", exp.Key, exp.Operator, slices.Sort(exp.Values)))
		}
	}
	return strings.Join(lines, "\n")
}
